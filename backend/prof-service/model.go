package main

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	Host  Role = "HOST"
	Guest Role = "GUEST"
)

type User struct {
	ID        string   `bson:"_id,omitempty" json:"userId" required:"true"`
	Username  string   `bson:"username,omitempty" json:"username" required:"true"`
	Email     string   `bson:"email,omitempty" json:"email" required:"true"`
	Role      Role     `bson:"role,omitempty" json:"role" `
	FirstName string   `bson:"firstname,omitempty" json:"firstname" required:"true"`
	LastName  string   `bson:"lastname,omitempty" json:"lastname" required:"true"`
	Location  Location `bson:"location,omitempty,inline" json:"location"`
}

type Location struct {
	Country      string `bson:"country,omitempty" json:"country" required:"true"`
	City         string `bson:"city,omitempty" json:"city" required:"true"`
	StreetName   string `bson:"streetName,omitempty" json:"streetName" required:"true"`
	StreetNumber string `bson:"streetNumber,omitempty" json:"streetNumber" required:"true"`
}

type ResponseUser struct {
	Username  string   `bson:"username,omitempty" json:"username"`
	Email     string   `bson:"email,omitempty" json:"email"`
	Role      Role     `bson:"role,omitempty" json:"role"`
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

func ValidateUser(user *User) error {
	if user.Username == "" {
		return errors.New("username is required")
	}

	if user.Email == "" {
		return errors.New("email is required")
	}

	if user.FirstName == "" {
		return errors.New("firsName is required")
	}

	if user.LastName == "" {
		return errors.New("lastName is required")
	}

	if user.Location.City == "" {
		return errors.New("city is required")
	}

	if user.Location.Country == "" {
		return errors.New("country is required")
	}

	if user.Location.StreetName == "" {
		return errors.New("streetName is required")
	}

	if user.Location.StreetNumber == "" {
		return errors.New("streetNumber is required")
	}

	return nil
}
