package models

import "time"

type LockSeatsRequest struct {
	ShowID      int   `json:"show_id"`
	ShowSeatIDs []int `json:"show_seat_ids"`
}

type ConfirmBookingRequest struct {
	BookingID     int    `json:"booking_id"`
	PaymentMethod string `json:"payment_method"`
}

type Booking struct {
	ID         int       `json:"id"`
	ShowID     int       `json:"show_id"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type BookingDetail struct {
	ID          int       `json:"id"`
	MovieTitle  string    `json:"movie_title"`
	TheaterName string    `json:"theater_name"`
	ScreenName  string    `json:"screen_name"`
	StartTime   time.Time `json:"start_time"`
	TotalPrice  float64   `json:"total_price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	Seats       []string  `json:"seats"`
}
