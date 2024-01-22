package main

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"time"
)

type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	HostId      string             `bson:"hostId, omitempty" json:"hostId"`
	Description string             `bson:"description,omitempty" json:"description"`
	Date        time.Time          `bson:"date,omitempty" json:"date"`
}

type Notifications []*Notification

type ReqToken struct {
	Token string `json:"token"`
}

func (as *Notifications) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(as)
}

func (a *Notification) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Notification) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}
