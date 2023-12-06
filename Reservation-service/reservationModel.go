package main

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gocql/gocql"
)

type ReservationByAccommodation struct {
	AccoId        gocql.UUID
	ReservationId gocql.UUID
	Price         int
	Date          time.Time
	IsDeleted     boolean
}

type ReservationByUser struct {
	UserId        gocql.UUID
	ReservationId gocql.UUID
	Price         int
	Date          time.Time
	IsDeleted     boolean
}

type ReservationsByAccommodation []*ReservationByAccommodation
type ReservationsByUser []*ReservationByUser

func (o *ReservationByAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
func (o *ReservationsByAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
func (o *ReservationsByUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
func (o *ReservationByUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *ReservationByAccommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
func (o *ReservationByUser) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
