package models

import "time"

type Show struct {
	ID        int       `json:"id"`
	MovieID   int       `json:"movie_id"`
	ScreenID  int       `json:"screen_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	BasePrice float64   `json:"base_price"`
}

type PostShowRequest struct {
	MovieID   int     `json:"movie_id"`
	ScreenID  int     `json:"screen_id"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	BasePrice float64 `json:"base_price"`
}

type ShowResponse struct {
	ID          int       `json:"id"`
	MovieTitle  string    `json:"movie_title"`
	TheaterName string    `json:"theater_name"`
	ScreenName  string    `json:"screen_name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	BasePrice   float64   `json:"base_price"`
}

type ShowSeatResponse struct {
	ShowSeatID int    `json:"show_seat_id"`
	RowChar    string `json:"row_char"`
	SeatNumber int    `json:"seat_number"`
	SeatType   string `json:"seat_type"`
	Status     string `json:"status"`
}
