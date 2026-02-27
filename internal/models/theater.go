package models

type Theater struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type PostTheaterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}
