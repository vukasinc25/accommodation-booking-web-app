package main

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gocql/gocql"
)

type ReservationByAccommodation struct {
	AccoId        gocql.UUID
	UserId        gocql.UUID
	ReservationId gocql.UUID
	NumberPeople  int
	StartDate     time.Time
	EndDate       time.Time
	IsDeleted     bool
}

type ReservationByUser struct {
	UserId        gocql.UUID
	AccoId        gocql.UUID
	ReservationId gocql.UUID
	NumberPeople  int
	StartDate     time.Time
	EndDate       time.Time
	IsDeleted     bool
}

type ReservationDateByAccomodationId struct {
	AccoId                string    `json:"acco_id"`
	BeginAccomodationDate time.Time `json:"begin_accomodation_date"`
	EndAccomodationDate   time.Time `json:"end_accomodation_date"`
}
type ReservationDate struct {
	BeginAccomodationDate time.Time `json:"begin_accomodation_date"`
	EndAccomodationDate   time.Time `json:"end_accomodation_date"`
}

type ReservationsByAccommodation []*ReservationByAccommodation
type ReservationsByUser []*ReservationByUser
type ReservationDatesByAccomodationId []*ReservationDate

func (o *ReservationByAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
func (o *ReservationDatesByAccomodationId) ToJSON(w io.Writer) error {
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
