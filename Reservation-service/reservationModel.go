package main

import (
	"time"
)

type Location struct {
	ID     int       `json:"ID"`
	Date   time.Time `json:"date"`
	accoId int       `json:"accoId"`
	userId int       `json:"userId"`
}
