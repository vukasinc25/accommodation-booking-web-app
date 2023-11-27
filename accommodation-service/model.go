package main

import (
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Accommodation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name      string             `bson:"name,omitempty" json:"name"`
	Location  Location           `bson:"location,omitempty,inline" json:"location"`
	Amenities []Amenity          `bson:"amenities,omitempty" json:"amenities"`
	MinGuests int                `bson:"minGuests,omitempty" json:"minGuests"`
	MaxGuests int                `bson:"maxGuests,omitempty" json:"maxGuests"`
	// Availability bool   `json:"availability"`
	// Details      string `json:"details"`
	Price     string 			`bson:"price,omitempty" json:"price"`
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
	HEATING          Amenity = "HEATING"
	AIR_CONDITIONING Amenity = "AIR_CONDITIONING"
	KITCHEN          Amenity = "KITCHEN"
	TV               Amenity = "TV"
	WASHER           Amenity = "WASHER"
)

type Accommodations []*Accommodation

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
