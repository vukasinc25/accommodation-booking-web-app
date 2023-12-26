package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	Host  Role = "HOST"
	Guest Role = "GUEST"
)

type User struct {
	ID        string   `bson:"_id,omitempty" json:"_id,omitempty"`
	Username  string   `bson:"username,omitempty" json:"username"`
	Email     string   `bson:"email,omitempty" json:"email"`
	Role      Role     `bson:"role,omitempty" json:"role"`
	FirstName string   `bson:"firstname,omitempty" json:"firstname"`
	LastName  string   `bson:"lastname,omitempty" json:"lastname"`
	Location  Location `bson:"location,omitempty,inline" json:"location"`
}

type Location struct {
	Country      string `bson:"country,omitempty" json:"country"`
	City         string `bson:"city,omitempty" json:"city"`
	StreetName   string `bson:"streetName,omitempty" json:"streetName"`
	StreetNumber string `bson:"streetNumber,omitempty" json:"streetNumber"`
}

type ResponseUser struct {
	Username  string   `bson:"username,omitempty" json:"username"`
	Email     string   `bson:"email,omitempty" json:"email"`
	FirstName string   `bson:"firstname,omitempty" json:"firstname"`
	LastName  string   `bson:"lastname,omitempty" json:"lastname"`
	Location  Location `bson:"location,omitempty,inline" json:"location"`
}

type VerifyEmail struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Username   string             `bson:"username,omitempty" json:"username" validate:"required"`
	Email      string             `bson:"email,omitempty" json:"email" validate:"required"`
	SecretCode string             `bson:"secretCode,omitempty" json:"secretCode" validate:"required"`
	IsUsed     bool               `bson:"isUsed" json:"isUsed" validate:"required"`
	CreatedAt  time.Time          `bson:"createdAt,omitempty" json:"createdAt" validate:"required"`
	ExpiredAt  time.Time          `bson:"expiredAt,omitempty" json:"expiredAt" validate:"required"`
}

type ReqToken struct {
	Token string `json:"token"`
}

type RequestId struct {
	UserId string `json:"userId"`
}

type Users []*User
