package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"movie-reservation-system/internal/middleware"
	"movie-reservation-system/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingHandler struct {
	DB *pgxpool.Pool
}

func (h *BookingHandler) LockSeatsHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.LockSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusInternalServerError)
		return
	}

	if len(req.ShowSeatIDs) == 0 {
		http.Error(w, "You must select at least one seat", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var bookingID int
	bookingQuery := `INSERT INTO bookings (user_id, show_id, total_price, status) VALUES ($1, $2, 0, 'pending') RETURNING id`
	err = tx.QueryRow(r.Context(), bookingQuery, claims.UserID, req.ShowID).Scan(&bookingID)
	if err != nil {
		log.Printf("Failed to create pending booking: %v", err)
		http.Error(w, "Failed to initiate booking", http.StatusInternalServerError)
		return
	}

	lockQuery := `
        UPDATE show_seats 
        SET 
            status = 'locked', 
            locked_until = NOW() + INTERVAL '5 minutes',
            booking_id = $1
        WHERE id = ANY($2) 
          AND show_id = $3
          AND (status = 'available' OR (status = 'locked' AND locked_until < NOW()))
        RETURNING id
    `

	rows, err := tx.Query(r.Context(), lockQuery, bookingID, req.ShowSeatIDs, req.ShowID)
	if err != nil {
		log.Printf("Error locking seats: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lockedCount int
	for rows.Next() {
		lockedCount++
	}

	if lockedCount != len(req.ShowSeatIDs) {
		http.Error(w, "One or more selected seats are no longer available. Please try again.", http.StatusConflict)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("Failed to commit seat lock transaction: %v", err)
		http.Error(w, "Failed to lock seats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Seats successfully locked for 5 minutes",
		"booking_id": bookingID,
	})
}

func (h *BookingHandler) ConfirmBookingHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ConfirmBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	checkQuery := `
        SELECT COUNT(ss.id), sh.base_price
        FROM show_seats ss
        JOIN shows sh ON ss.show_id = sh.id
        WHERE ss.booking_id = $1 AND ss.status = 'locked' AND ss.locked_until >= NOW()
        GROUP BY sh.base_price
    `

	var seatCount int
	var basePrice float64
	err = tx.QueryRow(r.Context(), checkQuery, req.BookingID).Scan(&seatCount, &basePrice)
	if err != nil {
		log.Printf("Failed to verify locked seats (might be expired): %v", err)
		http.Error(w, "Booking expired or invalid. Please select your seats again.", http.StatusBadRequest)
		return
	}

	totalPrice := float64(seatCount) * basePrice

	updateSeatsQuery := `
        UPDATE show_seats 
        SET status = 'booked', locked_until = NULL 
        WHERE booking_id = $1 AND status = 'locked'
    `
	_, err = tx.Exec(r.Context(), updateSeatsQuery, req.BookingID)
	if err != nil {
		log.Printf("Failed to update seat status to booked: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	updateBookingQuery := `
        UPDATE bookings 
        SET status = 'confirmed', total_price = $1 
        WHERE id = $2 AND user_id = $3 AND status = 'pending'
    `
	res, err := tx.Exec(r.Context(), updateBookingQuery, totalPrice, req.BookingID, claims.UserID)
	if err != nil {
		log.Printf("Failed to confirm booking record: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if res.RowsAffected() == 0 {
		http.Error(w, "Unauthorized or invalid booking", http.StatusUnauthorized)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("Failed to commit confirmation: %v", err)
		http.Error(w, "Failed to finalize booking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Payment successful! Booking confirmed.",
		"booking_id":   req.BookingID,
		"total_amount": totalPrice,
		"tickets":      seatCount,
	})
}

func (h *BookingHandler) GetBookingsHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `SELECT id, show_id, total_price, status, created_at FROM bookings WHERE user_id = $1`
	rows, err := h.DB.Query(r.Context(), query, claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	bookings := []models.Booking{}
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.ShowID, &b.TotalPrice, &b.Status, &b.CreatedAt); err != nil {
			http.Error(w, "Failed to fetch booking history", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, b)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating bookings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookings)
}

func (h *BookingHandler) GetBookingByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var ticket models.BookingDetail
	ticket.Seats = []string{}

	ticketQuery := `
        SELECT 
            b.id, m.title, t.name, s.name, sh.start_time, 
            b.total_price, b.status, b.created_at
        FROM bookings b
        JOIN shows sh ON b.show_id = sh.id
        JOIN movies m ON sh.movie_id = m.id
        JOIN screens s ON sh.screen_id = s.id
        JOIN theaters t ON s.theater_id = t.id
        WHERE b.id = $1 AND b.user_id = $2
    `

	err = h.DB.QueryRow(r.Context(), ticketQuery, id, claims.UserID).Scan(
		&ticket.ID, &ticket.MovieTitle, &ticket.TheaterName,
		&ticket.ScreenName, &ticket.StartTime, &ticket.TotalPrice,
		&ticket.Status, &ticket.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error fetching ticket details for booking %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	seatQuery := `
        SELECT s.row_char, s.seat_number 
        FROM show_seats ss
        JOIN seats s ON ss.seat_id = s.id
        WHERE ss.booking_id = $1
        ORDER BY s.row_char, s.seat_number
    `
	rows, err := h.DB.Query(r.Context(), seatQuery, id)
	if err != nil {
		log.Printf("Error querying seats for booking %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var rowChar string
		var seatNum int
		if err := rows.Scan(&rowChar, &seatNum); err != nil {
			log.Printf("Error scanning seat row: %v", err)
			http.Error(w, "Error parsing seats", http.StatusInternalServerError)
			return
		}
		seatLabel := fmt.Sprintf("%s%d", rowChar, seatNum)
		ticket.Seats = append(ticket.Seats, seatLabel)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating seat rows: %v", err)
		http.Error(w, "Error processing seats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

func (h *BookingHandler) CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	bookingID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		log.Printf("Failed to begin transaction for cancellation: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var currentStatus string
	var startTime time.Time

	validationQuery := `
        SELECT b.status, sh.start_time 
        FROM bookings b
        JOIN shows sh ON b.show_id = sh.id
        WHERE b.id = $1 AND b.user_id = $2
    `
	err = tx.QueryRow(r.Context(), validationQuery, bookingID, claims.UserID).Scan(&currentStatus, &startTime)
	if err == pgx.ErrNoRows {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error validating booking %d for cancellation: %v", bookingID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if currentStatus == "cancelled" {
		http.Error(w, "Booking is already cancelled", http.StatusBadRequest)
		return
	}
	if time.Now().After(startTime) {
		http.Error(w, "Cannot cancel a ticket for a show that has already started", http.StatusBadRequest)
		return
	}

	updateBookingQuery := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
	_, err = tx.Exec(r.Context(), updateBookingQuery, bookingID)
	if err != nil {
		log.Printf("Failed to update booking status: %v", err)
		http.Error(w, "Failed to cancel booking", http.StatusInternalServerError)
		return
	}

	freeSeatsQuery := `
        UPDATE show_seats 
        SET status = 'available', booking_id = NULL, locked_until = NULL 
        WHERE booking_id = $1
    `
	_, err = tx.Exec(r.Context(), freeSeatsQuery, bookingID)
	if err != nil {
		log.Printf("Failed to release seats for booking %d: %v", bookingID, err)
		http.Error(w, "Failed to release seats", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("Failed to commit cancellation transaction: %v", err)
		http.Error(w, "Failed to finalize cancellation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Booking successfully cancelled. Your seats have been released.",
	})
}
