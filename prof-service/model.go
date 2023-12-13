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
	ID       string `bson:"_id,omitempty" json:"_id,omitempty"`
	Username string `bson:"username,omitempty" json:"username"`
	Role     Role   `bson:"role,omitempty" json:"role"`
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

type Users []*User
