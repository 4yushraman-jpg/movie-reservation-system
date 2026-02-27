package models

type Screen struct {
	ID         int    `json:"id"`
	TheaterID  int    `json:"theater_id"`
	Name       string `json:"name"`
	TotalSeats int    `json:"total_seats"`
}

type PostScreenRequest struct {
	Name       string `json:"name"`
	TotalSeats int    `json:"total_seats"`
}

type GetScreenResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TotalSeats int    `json:"total_seats"`
}
