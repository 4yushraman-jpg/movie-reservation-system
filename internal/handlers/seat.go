package handlers

import (
	"encoding/json"
	"log"
	"movie-reservation-system/internal/models"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeatHandler struct {
	DB *pgxpool.Pool
}

func (h *SeatHandler) GenerateSeatsHandler(w http.ResponseWriter, r *http.Request) {
	screenIDStr := chi.URLParam(r, "screen_id")
	screenID, err := strconv.Atoi(screenIDStr)
	if err != nil {
		http.Error(w, "Invalid screen ID", http.StatusBadRequest)
		return
	}

	var req models.GenerateSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.RowLabels) == 0 || req.SeatsPerRow <= 0 || req.SeatType == "" {
		http.Error(w, "Row labels can't be empty, seats per row must be more than zero and seat type mustn't be empty", http.StatusBadRequest)
		return
	}

	var rows [][]interface{}
	for _, rowChar := range req.RowLabels {
		for seatNum := 1; seatNum <= req.SeatsPerRow; seatNum++ {
			rows = append(rows, []interface{}{screenID, rowChar, seatNum, req.SeatType})
		}
	}

	copyCount, err := h.DB.CopyFrom(
		r.Context(),
		pgx.Identifier{"seats"},
		[]string{"screen_id", "row_char", "seat_number", "seat_type"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		log.Printf("Failed to bulk insert seats for screen %d: %v", screenID, err)
		http.Error(w, "Failed to generate seats. They might already exist.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "seats generated successfully",
		"seats_created": copyCount,
	})
}
