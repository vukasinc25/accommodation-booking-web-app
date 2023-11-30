package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	utility "github.com/vukasinc25/fst-airbnb/utility/messaging"
	nats2 "github.com/vukasinc25/fst-airbnb/utility/messaging/nats"

	"github.com/gorilla/mux"

	"github.com/thanhpk/randstr"
	"github.com/vukasinc25/fst-airbnb/mail"
	"github.com/vukasinc25/fst-airbnb/token"
)

// UserHandler handles HTTP requests related to user operations.
type UserHandler struct {
	logger   *log.Logger
	db       *UserRepo
	jwtMaker token.Maker
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(l *log.Logger, r *UserRepo, jwtMaker token.Maker) *UserHandler {
	return &UserHandler{l, r, jwtMaker}
}

func InitPubSub() utility.Subscriber {
	subscriber, err := nats2.NewNATSSubscriber("auth.check")
	if err != nil {
		log.Fatal(err)
	}
	return subscriber
}
func (uh *UserHandler) Auth(msg *nats.Msg) nats.Msg {

	uh.logger.Println("Received publish")

	uh.logger.Println(string(msg.Data))

	//payload, _ := uh.jwtMaker.VerifyToken(string(msg.Data))
	//uh.logger.Println(payload.Role)
	//if err != nil || payload.Role != "HOST" {
	//	msg2 := nats.Msg{Data: []byte("not ok")}
	//	return msg2
	//} //TODO fix

	msg2 := nats.Msg{Data: []byte("ok")}
	return msg2
}

// createUser handles user creation requests.
func (uh *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Println("Error cant mimi.ParseMediaType")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	log.Println("Pre decodeBody")
	rt, err := decodeBody(req.Body)
	if err != nil {
		if strings.Contains(err.Error(), "Key: 'User.Username' Error:Field validation for 'Username' failed on the 'min' tag") {
			http.Error(w, "Username must have minimum 6 characters", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Key: 'User.Password' Error:Field validation for 'Password' failed on the 'password' tag") {
			http.Error(w, "Password must have minimum 8 characters,minimum one big letter, numbers and special characters", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag") {
			http.Error(w, "Email format is incorrect", http.StatusBadRequest)
		} else {
			http.Error(w, "Ovde"+err.Error(), http.StatusBadRequest)
		}
		return
	}

	rt.IsEmailVerified = false

	sanitizedUsername := sanitizeInput(rt.Username)
	sanitizedPassword := sanitizeInput(rt.Password)
	sanitizedRole := sanitizeInput(string(rt.Role))

	rt.Username = sanitizedUsername
	rt.Password = sanitizedPassword
	rt.Role = Role(sanitizedRole)

	// Fetch the blacklist
	blacklist, err := NewBlacklistFromURL()
	if err != nil {
		log.Println("Error fetching blacklist: %v\n", err)
		return
	}

	log.Println(sanitizedUsername)
	log.Println(sanitizedPassword)
	log.Println(sanitizedRole)

	// Check if the password is blacklisted
	if blacklist.IsBlacklisted(rt.Password) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Password is not good")
		return
	}

	log.Println("Not hashed Password: %w", rt.Password)
	// Hash the password before storing
	hashedPassword, err := HashPassword(rt.Password)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	rt.Password = hashedPassword
	log.Println("Hashed Password: %w", rt.Password)

	uh.sendEmail(rt)

	err = uh.db.Insert(rt)
	if err != nil {
		if strings.Contains(err.Error(), "username") {
			http.Error(w, "Provide different username", http.StatusConflict)
		} else if strings.Contains(err.Error(), "email") {
			http.Error(w, "Provide different email", http.StatusConflict)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

// getAllUsers handles requests to retrieve all users.
func (uh *UserHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {
	// Retrieve all users from the database
	users, err := uh.db.GetAll()
	ctx := req.Context()

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	if users == nil {
		return
	}

	// Retrieve the authorization payload from the request context
	authPayload, ok := ctx.Value(AuthorizationPayloadKey).(*token.Payload)
	if !ok || authPayload == nil {
		http.Error(w, "Authorization payload not found", http.StatusInternalServerError)
		return
	}

	// Check user role for authorization
	if authPayload.Role == "guest" {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	// Convert users to JSON and send the response
	err = users.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to JSON", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to JSON:", err)
		return
	}
}

// loginUser handles user login requests.
func (uh *UserHandler) loginUser(w http.ResponseWriter, req *http.Request) {
	rt, err := decodeLoginBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := rt.Username
	password := rt.Password
	user, err := uh.db.GetByUsername(username)
	if err != nil {
		log.Println("mongo: no documents in result: treba da se registuje neko")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prooveravamo da li korisnik ima verifikovan mejl 169,170,171,172,173,174
	log.Println(user.IsEmailVerified)
	if !user.IsEmailVerified {
		http.Error(w, "Morate da verifikujete email i treba da postoji dugme da se ponovo posalje email ili da mu se napise da proveri email ako nije", http.StatusBadRequest)
		return
	}

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	// If user is not found, return an error
	if user == nil {
		http.Error(w, "Invalid username or password", http.StatusNotFound)
		return
	}

	// Check if the provided password matches the hashed password in the database
	err = CheckHashedPassword(password, user.Password)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate and send JWT token as a response
	jwtToken(user, w, uh)
}

func (uh *UserHandler) sendEmail(newUser *User) error {
	log.Println("SendEmail()")

	randomCode := randstr.String(20)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	verificationEmail := VerifyEmail{
		Username:   newUser.Username,
		Email:      newUser.Email,
		SecretCode: randomCode,
		IsUsed:     false,
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(15 * time.Minute), // moras da promenis da je trajanje 15 min
	}

	err := uh.db.CreateVerificationEmail(verificationEmail)
	if err != nil {
		log.Println("Cant save verification email in SendEmail()method")
		return err
	}

	sender := mail.NewGmailSender("Air Bnb", "mobilneaplikcijesit@gmail.com", "esrqtcomedzeapdr", tlsConfig) //postavi recoveri password
	subject := "A test email"
	content := fmt.Sprintf(`
    <h1>Hello world</h1>
    <h1>This is a test message from AirBnb</h1>
    <h4>Authorization code for password change: %s</h4>`, randomCode)
	to := []string{"mobilneaplikcijesit@gmail.com"}
	attachFiles := []string{}
	log.Println("Pre SendEmail(subject, content, to, nil, nil, attachFiles)")
	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Cant send email")
		return err
	}

	return nil
	// w.WriteHeader(http.StatusCreated)
	// message := "Poslat je mail na moblineaplikacijesit@gmail.com"
	// renderJSON(w, message)
}

func (uh *UserHandler) ChangePassword(w http.ResponseWriter, req *http.Request) {

}

func (uh *UserHandler) verifyEmail(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	code := vars["code"]

	verificationEmail, err := uh.db.GetVerificationEmailByCode(code)
	if err != nil {
		log.Println("Error in getting verificationEmail:", err)
		http.Error(w, "Error in getting verificationEmail", http.StatusInternalServerError)
		return
	}

	if verificationEmail != nil {
		if !verificationEmail.IsUsed {
			isActive, err := uh.db.IsVerificationEmailActive(code)
			if err != nil {
				log.Println("Error Verification code is not active")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if isActive {
				err = uh.db.UpdateUsersVerificationEmail(verificationEmail.Username)
				if err != nil {
					log.Println("Error in trying to update UsersVerificationEmail")
					fmt.Println("Error in trying to update UsersVerificationEmail")
					return
				}

				err = uh.db.UpdateVerificationEmail(code)
				if err != nil {
					log.Println("Error in trying to update VerificationEmail")
					fmt.Println("Error in trying to update VerificationEmail")
					return
				}
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("Your mail have been verified"))
			} else {
				http.Error(w, "Code is not active", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Code that has been forwarded has been used", http.StatusBadRequest)
			return
		}
	}

}
func jwtToken(user *User, w http.ResponseWriter, uh *UserHandler) {
	durationStr := "15m" // Should be a constant outside the function
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Println("Cannot parse duration")
		return
	}

	accessToken, accessPayload, err := uh.jwtMaker.CreateToken(
		user.ID,
		user.Username,
		string(user.Role),
		duration,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp := LoginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	e := json.NewEncoder(w)
	e.Encode(rsp)
}

// decodeBody decodes the request body into a User struct.
func decodeBody(r io.Reader) (*User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt User
	if err := dec.Decode(&rt); err != nil {
		log.Println("Decode cant be done")
		return nil, err
	}

	if err := ValidateUser(rt); err != nil {
		log.Println("User is not succesfuly validated in ValidateUser func")
		return nil, err
	}

	return &rt, nil
}

// decodeLoginBody decodes the request body into a LoginUser struct.
func decodeLoginBody(r io.Reader) (*LoginUser, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt LoginUser
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

// sanitizeInput replaces "<" with "&lt;" to prevent potential HTML/script injection.
func sanitizeInput(input string) string {
	sanitizedInput := strings.ReplaceAll(input, "<", "&lt;")
	return sanitizedInput
}

// renderJSON writes JSON data to the response writer.
func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// ToJSON converts a Users object to JSON and writes it to the response writer.
func (u *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

// ToJSON converts a User object to JSON and writes it to the response writer.
func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}
