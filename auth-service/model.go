package main

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

type Users []*User
