package handlers

import (
	"encoding/json"
	"log"
	"movie-reservation-system/internal/models"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TheaterHandler struct {
	DB *pgxpool.Pool
}

func (h *TheaterHandler) PostTheaterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PostTheaterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Location == "" {
		http.Error(w, "Name and location cannot be empty", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO theaters (name, location) VALUES ($1, $2) RETURNING id`

	var insertedID int
	err := h.DB.QueryRow(r.Context(), query, req.Name, req.Location).Scan(&insertedID)
	if err != nil {
		log.Printf("Failed to insert theater: %v", err) // Log the raw error
		http.Error(w, "Failed to save theater", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Theater added successfully",
		"theater_id": insertedID,
	})
}

func (h *TheaterHandler) GetTheatersHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, location FROM theaters`
	rows, err := h.DB.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	theaters := []models.Theater{}

	for rows.Next() {
		var t models.Theater
		if err := rows.Scan(&t.ID, &t.Name, &t.Location); err != nil {
			http.Error(w, "Failed to fetch theaters", http.StatusInternalServerError)
			return
		}
		theaters = append(theaters, t)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating theaters", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(theaters)
}
