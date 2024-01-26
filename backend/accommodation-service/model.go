package main

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Accommodation struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name         string             `bson:"name,omitempty" json:"name"`
	Location     Location           `bson:"location,omitempty,inline" json:"location"`
	Amenities    []Amenity          `bson:"amenities,omitempty" json:"amenities"`
	MinGuests    int                `bson:"minGuests,omitempty" json:"minGuests"`
	MaxGuests    int                `bson:"maxGuests,omitempty" json:"maxGuests"`
	Username     string             `bson:"username,omitempty" json:"username"`
	AverageGrade float64            `bson:"averageGrade,omitempty" json:"averageGrade`
<<<<<<< Updated upstream
	Images       []string           `bson:"images,omitempty" json:"images"`
=======
>>>>>>> Stashed changes
	// Availability bool   `json:"availability"`
	// Details      string `json:"details"`
	//Price     string 			`bson:"price,omitempty" json:"price"`
}

type UserReservations struct {
	ReservationId  string    `json:"reservationId" validate:"required"`
	AccoId         string    `json:"accoId" validate:"required"`
	Price          int       `json:"price" validate:"required"`
	StartDate      time.Time `json:"startDate" validate:"required"`
	NumberOfPeople int       `json:"numberOfPeople"`
	EndDate        time.Time `json:"endDate" validate:"required"`
}

type Location struct {
	Country      string `bson:"country,omitempty" json:"country"`
	City         string `bson:"city,omitempty" json:"city"`
	StreetName   string `bson:"streetName,omitempty" json:"streetName"`
	StreetNumber string `bson:"streetNumber,omitempty" json:"streetNumber"`
}

type Amenity string

const (
	WIFI             Amenity = "WIFI"
	Heating          Amenity = "Heating"
	Air_conditioning Amenity = "Air conditioning"
	Kitchen          Amenity = "Kitchen"
	TV               Amenity = "TV"
	Washer           Amenity = "Washer"
)

type Accommodations []*Accommodation
type AccommodationGrades []*AccommodationGrade
type ReservationsByUser []*UserReservations

type ReqToken struct {
	Token string `json:"token"`
}

type AccommodationGrade struct {
	ID              string `bson:"_id,omitempty" json:"id"`
	UserId          string `bson:"userId,omitempty" json:"userId"`
	AccommodationId string `bson:"accommodationId,omitempty" json:"accommodationId"`
	CreatedAt       string `bson:"createdAt,omitempty" json:"createdAt"`
	Grade           int    `bson:"grade,omitempty" json:"grade"`
}

func (as *Accommodations) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(as)
}

func (a *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}

func (a *AccommodationGrades) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func ValidateAccommodationGrade(accommodationGrade *AccommodationGrade) error {
	// if accommodationGrade.UserId == "" {
	// 	return errors.New("userId is required")
	// }
	if accommodationGrade.AccommodationId == "" {
		return errors.New("hostId is required")
	}
	if accommodationGrade.Grade == 0 {
		return errors.New("grade is required")
	}
	if accommodationGrade.Grade <= 0 || accommodationGrade.Grade > 5 {
		return errors.New("grade must be in the range of 1 to 5")
	}

	return nil
}
