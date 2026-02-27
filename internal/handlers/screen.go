package handlers

import (
	"encoding/json"
	"log"
	"movie-reservation-system/internal/models"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScreenHandler struct {
	DB *pgxpool.Pool
}

func (h *ScreenHandler) PostScreenHandler(w http.ResponseWriter, r *http.Request) {
	theaterIDStr := chi.URLParam(r, "theater_id")
	theaterID, err := strconv.Atoi(theaterIDStr)
	if err != nil {
		http.Error(w, "Invalid theater ID", http.StatusBadRequest)
		return
	}

	var req models.PostScreenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.TotalSeats <= 0 {
		http.Error(w, "Name cannot be empty and total seats must be greater than zero", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO screens (theater_id, name, total_seats) VALUES ($1, $2, $3) RETURNING id`

	var insertedID int
	err = h.DB.QueryRow(r.Context(), query, theaterID, req.Name, req.TotalSeats).Scan(&insertedID)
	if err != nil {
		log.Printf("Failed to insert screen for theater %d: %v", theaterID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Screen added successfully",
		"screen_id": insertedID,
	})
}

func (h *ScreenHandler) GetScreensHandler(w http.ResponseWriter, r *http.Request) {
	theaterIDStr := chi.URLParam(r, "theater_id")
	theaterID, err := strconv.Atoi(theaterIDStr)
	if err != nil {
		http.Error(w, "Invalid theater ID", http.StatusBadRequest)
		return
	}

	query := `SELECT id, name, total_seats FROM screens WHERE theater_id = $1`
	rows, err := h.DB.Query(r.Context(), query, theaterID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	screens := []models.GetScreenResponse{}

	for rows.Next() {
		var s models.GetScreenResponse
		if err := rows.Scan(&s.ID, &s.Name, &s.TotalSeats); err != nil {
			http.Error(w, "Failed to fetch screens", http.StatusInternalServerError)
			return
		}
		screens = append(screens, s)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating screens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(screens)
}
