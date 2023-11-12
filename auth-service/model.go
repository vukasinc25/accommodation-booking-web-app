package main

import (
	"time"
)

type User struct {
	ID       string `bson:"_id,omitempty" json:"_id,omitempty"`
	Username string `bson:"username,omitempty" json:"username"`
	Password string `bson:"password,omitempty" json:"password"`
	Role     string `bson:"role,omitempty" json:"role"`
	// HashedPassword string `bson:"hashed_password,omitempty" json:"hashed_password"`
	// FullName       string `bson:"fullname,omitempty" json:"fullname"`
	// Email          string `bson:"email,omitempty" json:"email"`
	// CreatedAt      string `bson:"created_at,omitempty" json:"created_at"`
}

type ResponseUser struct {
	ID       string `bson:"_id,omitempty" json:"_id,omitempty"`
	Username string `bson:"username,omitempty" json:"username"`
	Role     string `bson:"role,omitempty" json:"role"`
	// HashedPassword string `bson:"hashed_password,omitempty" json:"hashed_password"`
	// FullName       string `bson:"fullname,omitempty" json:"fullname"`
	// Email          string `bson:"email,omitempty" json:"email"`
	// CreatedAt      string `bson:"created_at,omitempty" json:"created_at"`
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type Users []*ResponseUser
