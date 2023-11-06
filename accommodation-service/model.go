package main

type Accommodation struct {
	Name      string `json:"name"`
	Details   string `json:"details"`
	MinGuests int    `json:"min_guests"`
	MaxGuests int    `json:"max_guests"`
}