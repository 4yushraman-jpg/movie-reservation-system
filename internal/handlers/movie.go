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

type MovieHandler struct {
	DB *pgxpool.Pool
}

func (h *MovieHandler) PostMovieHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PostMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.DurationMinutes <= 0 {
		http.Error(w, "Title cannot be empty and duration must be greater than zero", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		http.Error(w, "Invalid release date format. Please use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	query := `
        INSERT INTO movies (title, description, duration_minutes, release_date) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id
    `

	var insertedID int
	err = h.DB.QueryRow(r.Context(), query, req.Title, req.Description, req.DurationMinutes, parsedDate).Scan(&insertedID)
	if err != nil {
		log.Printf("Failed to insert movie: %v", err)
		http.Error(w, "Failed to save movie to the database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Movie added successfully",
		"movie_id": insertedID,
	})
}

func (h *MovieHandler) PutMovieHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var req models.PostMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.DurationMinutes <= 0 {
		http.Error(w, "Title cannot be empty and duration must be greater than zero", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		http.Error(w, "Invalid release date format. Please use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	query := `
        UPDATE movies 
        SET title = $1, description = $2, duration_minutes = $3, release_date = $4
        WHERE id = $5
    `

	res, err := h.DB.Exec(r.Context(), query, req.Title, req.Description, req.DurationMinutes, parsedDate, id)
	if err != nil {
		log.Printf("Failed to update movie %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if res.RowsAffected() == 0 {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Movie updated successfully",
	})
}

func (h *MovieHandler) DeleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM movies WHERE id = $1`

	res, err := h.DB.Exec(r.Context(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MovieHandler) GetMoviesHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, description, duration_minutes, release_date FROM movies`
	rows, err := h.DB.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	movies := []models.Movie{}

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.DurationMinutes, &m.ReleaseDate); err != nil {
			http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
			return
		}
		movies = append(movies, m)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating movies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movies)
}

func (h *MovieHandler) GetMovieByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	query := `SELECT id, title, description, duration_minutes, release_date FROM movies WHERE id = $1`
	var movie models.Movie
	err = h.DB.QueryRow(r.Context(), query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.DurationMinutes,
		&movie.ReleaseDate,
	)
	if err == pgx.ErrNoRows {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to fetch movie %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(movie)
}
