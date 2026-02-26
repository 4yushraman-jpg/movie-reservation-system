package models

import "time"

type Movie struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date"`
}

type PostMovieRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	ReleaseDate     string `json:"release_date"`
}
