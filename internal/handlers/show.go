package handlers

import (
	"encoding/json"
	"log"
	"movie-reservation-system/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShowHandler struct {
	DB *pgxpool.Pool
}

func (h *ShowHandler) PostShowHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PostShowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format. Use RFC3339 (e.g., 2026-03-01T18:00:00Z)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format.", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var showID int
	showQuery := `
        INSERT INTO shows (movie_id, screen_id, start_time, end_time, base_price) 
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `
	err = tx.QueryRow(r.Context(), showQuery, req.MovieID, req.ScreenID, startTime, endTime, req.BasePrice).Scan(&showID)
	if err != nil {
		log.Printf("Failed to insert show: %v", err)
		http.Error(w, "Failed to schedule show", http.StatusInternalServerError)
		return
	}

	seatQuery := `SELECT id FROM seats WHERE screen_id = $1`
	rows, err := tx.Query(r.Context(), seatQuery, req.ScreenID)
	if err != nil {
		log.Printf("Failed to fetch seats for screen %d: %v", req.ScreenID, err)
		http.Error(w, "Failed to generate show seats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seatIDs []int
	for rows.Next() {
		var seatID int
		if err := rows.Scan(&seatID); err != nil {
			log.Printf("Error scanning seat ID: %v", err)
			http.Error(w, "Error processing seats", http.StatusInternalServerError)
			return
		}
		seatIDs = append(seatIDs, seatID)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Error processing seats", http.StatusInternalServerError)
		return
	}

	var copyRows [][]interface{}
	for _, seatID := range seatIDs {
		copyRows = append(copyRows, []interface{}{showID, seatID, "available"})
	}

	copyCount, err := tx.CopyFrom(
		r.Context(),
		pgx.Identifier{"show_seats"},
		[]string{"show_id", "seat_id", "status"},
		pgx.CopyFromRows(copyRows),
	)
	if err != nil {
		log.Printf("Failed to bulk insert show_seats: %v", err)
		http.Error(w, "Failed to generate show seats", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Failed to save show", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Show scheduled successfully",
		"show_id":       showID,
		"tickets_ready": copyCount,
	})
}

func (h *ShowHandler) GetShowsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT 
            sh.id, 
            m.title, 
            t.name AS theater_name, 
            s.name AS screen_name, 
            sh.start_time, 
            sh.end_time, 
            sh.base_price
        FROM shows sh
        JOIN movies m ON sh.movie_id = m.id
        JOIN screens s ON sh.screen_id = s.id
        JOIN theaters t ON s.theater_id = t.id
        WHERE sh.start_time > NOW()
        ORDER BY sh.start_time ASC
    `

	rows, err := h.DB.Query(r.Context(), query)
	if err != nil {
		log.Printf("Error querying shows: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	shows := []models.ShowResponse{}

	for rows.Next() {
		var s models.ShowResponse
		if err := rows.Scan(
			&s.ID,
			&s.MovieTitle,
			&s.TheaterName,
			&s.ScreenName,
			&s.StartTime,
			&s.EndTime,
			&s.BasePrice,
		); err != nil {
			log.Printf("Error scanning show row: %v", err)
			http.Error(w, "Failed to fetch shows", http.StatusInternalServerError)
			return
		}
		shows = append(shows, s)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating show rows: %v", err)
		http.Error(w, "Error iterating shows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shows)
}

func (h *ShowHandler) GetShowSeatsHandler(w http.ResponseWriter, r *http.Request) {
	showIDStr := chi.URLParam(r, "show_id")
	showID, err := strconv.Atoi(showIDStr)
	if err != nil {
		http.Error(w, "Invalid show ID", http.StatusBadRequest)
		return
	}

	query := `
        SELECT 
            ss.id, 
            s.row_char, 
            s.seat_number, 
            s.seat_type, 
            ss.status
        FROM show_seats ss
        JOIN seats s ON ss.seat_id = s.id
        WHERE ss.show_id = $1
        ORDER BY s.row_char, s.seat_number
    `

	rows, err := h.DB.Query(r.Context(), query, showID)
	if err != nil {
		log.Printf("Error querying seats for show %d: %v", showID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	showSeats := []models.ShowSeatResponse{}

	for rows.Next() {
		var ss models.ShowSeatResponse
		if err := rows.Scan(&ss.ShowSeatID, &ss.RowChar, &ss.SeatNumber, &ss.SeatType, &ss.Status); err != nil {
			log.Printf("Error scanning show seat row: %v", err)
			http.Error(w, "Failed to fetch show seats", http.StatusInternalServerError)
			return
		}
		showSeats = append(showSeats, ss)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating show seat rows: %v", err)
		http.Error(w, "Error iterating show seats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(showSeats)
}
