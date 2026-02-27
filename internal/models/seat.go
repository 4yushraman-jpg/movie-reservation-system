package models

type Seat struct {
	ID         int    `json:"id"`
	ScreenID   int    `json:"screen_id"`
	RowChar    string `json:"row_char"`
	SeatNumber int    `json:"seat_number"`
	SeatType   string `json:"seat_type"`
}

type GenerateSeatsRequest struct {
	RowLabels   []string `json:"row_labels"`
	SeatsPerRow int      `json:"seats_per_row"`
	SeatType    string   `json:"seat_type"`
}
