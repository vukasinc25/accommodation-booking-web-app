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
	//Price     string 			`bson:"price,omitempty" json:"price"`
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

type ReqToken struct {
	Token string `json:"token"`
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
