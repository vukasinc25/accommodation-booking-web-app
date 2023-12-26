package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	Host  Role = "HOST"
	Guest Role = "GUEST"
)

type UserA struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Username        string             `bson:"username,omitempty" json:"username"`
	Password        string             `bson:"password,omitempty" json:"password"`
	Email           string             `bson:"email,omitempty" json:"email"`
	IsEmailVerified bool               `bson:"isEmailVerified" json:"isEmailVerified"`
	Role            Role               `bson:"role,omitempty" json:"role"`
}
type UserB struct {
	ID        string   `bson:"_id,omitempty" json:"_id,omitempty"`
	Username  string   `bson:"username,omitempty" json:"username"`
	Email     string   `bson:"email,omitempty" json:"email"`
	Role      Role     `bson:"role,omitempty" json:"role"`
	FirstName string   `bson:"firstname,omitempty" json:"firstname"`
	LastName  string   `bson:"lastname,omitempty" json:"lastname"`
	Location  Location `bson:"location,omitempty,inline" json:"location"`
}
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Username        string             `bson:"username,omitempty" json:"username" validate:"required,min=6"`
	Password        string             `bson:"password,omitempty" json:"password" validate:"required,password"`
	Role            Role               `bson:"role,omitempty" json:"role" validate:"required,oneof=HOST GUEST"`
	Email           string             `bson:"email,omitempty" json:"email" validate:"required,email"`
	IsEmailVerified bool               `bson:"isEmailVerified" json:"isEmailVerified"`
	FirstName       string             `bson:"firstname,omitempty" json:"firstname"`
	LastName        string             `bson:"lastname,omitempty" json:"lastname"`
	Location        Location           `bson:"location,omitempty,inline" json:"location"`
}

type Location struct {
	Country      string `bson:"country,omitempty" json:"country" validate:"required"`
	City         string `bson:"city,omitempty" json:"city" validate:"required"`
	StreetName   string `bson:"streetName,omitempty" json:"streetName" validate:"required"`
	StreetNumber string `bson:"streetNumber,omitempty" json:"streetNumber" validate:"required"`
}

type ResponseUser struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Username string             `bson:"username,omitempty" json:"username"`
	Role     string             `bson:"role,omitempty" json:"role"`
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

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type SiteVerifyRequest struct {
	RecaptchaResponse string `json:"g-recaptcha-response"`
}

type ForgottenPassword struct {
	NewPassword     string `bson:"newPassword,omitempty" json:"newPassword" validate:"required,newPassword"`
	ConfirmPassword string `bson:"confirmPassword,omitempty" json:"confirmPassword" validate:"required"`
	Code            string `bson:"code,omitempty" json:"code" validate:"required"`
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

type ForgottenPasswordEmail struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email      string             `bson:"email,omitempty" json:"email" validate:"required"`
	SecretCode string             `bson:"secretCode,omitempty" json:"secretCode" validate:"required"`
	IsUsed     bool               `bson:"isUsed" json:"isUsed" validate:"required"`
	CreatedAt  time.Time          `bson:"createdAt,omitempty" json:"createdAt" validate:"required"`
	ExpiredAt  time.Time          `bson:"expiredAt,omitempty" json:"expiredAt" validate:"required"`
}

type ReqToken struct {
	Token string `json:"token"`
}

func ValidateUser(user User) error {
	validate := validator.New()

	// Register custom validation tag for password complexity
	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		return len(password) >= 8 &&
			strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
			strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
			strings.ContainsAny(password, "0123456789") &&
			regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	})

	validate.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		return govalidator.IsEmail(email)
	})

	return validate.Struct(user)
}

func ValidateForgottenPassword(forgottenPassword ForgottenPassword) error {
	validate := validator.New()

	// Register custom validation tag for password complexity
	validate.RegisterValidation("newPassword", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		return len(password) >= 8 &&
			strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
			strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") &&
			strings.ContainsAny(password, "0123456789") &&
			regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	})

	return validate.Struct(forgottenPassword)
}
