package main

type Location struct {
	ID           int    `json:"ID"`
	Country      string `json:"country"`
	City         string `json:"city"`
	StreetName   string `json:"streetName"`
	StreetNumber string `json:"streetNumber"`
}
