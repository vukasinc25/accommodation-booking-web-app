package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func InitPubSubAuth() utility.Subscriber {
	subscriber, err := nats2.NewNATSSubscriber("auth.check")
	if err != nil {
		log.Fatal(err)
	}
	return subscriber
}

// func InitPubSubUsername() utility.Subscriber {
// 	subscriber, err := nats2.NewNATSSubscriber("prof.publish.username")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return subscriber
// }

// func (uh *UserHandler) Create(username string, password string, email string) error {
// 	log.Println("Usli u Create")
// 	user := Use{
// 		Username:        username,
// 		Password:        password,
// 		Email:           email,
// 		IsEmailVerified: false,
// 	}
// 	err := uh.db.InsertUser(&user)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (uh *UserHandler) SubscribeUsername(msg *nats.Msg) nats.Msg {
// 	log.Println("Received publish")

// 	// Decode the Base64 encoded username
// 	decodedUsername := msg.Data

// 	log.Println("Received username: ", string(decodedUsername))

// 	err := uh.db.UpdateUsersVerificationEmail("perapera")
// 	if err != nil {
// 		return nats.Msg{Data: []byte("Eror u izmeni usera")}
// 	}

// 	responseMsg := nats.Msg{Data: []byte("ok")}
// 	return responseMsg
// }

func (uh *UserHandler) Auth(msg *nats.Msg) nats.Msg {

	uh.logger.Println("Received publish")

	uh.logger.Println(string(msg.Data))
	uh.logger.Println(strings.Split(string(msg.Data), "\"")[3])
	jToken := strings.Split(string(msg.Data), "\"")[3]

	payload, err := uh.jwtMaker.VerifyToken(jToken)

	uh.logger.Println(payload.Role)
	if err != nil || payload.Role != "HOST" {
		msg2 := nats.Msg{Data: []byte("not ok")}
		return msg2
	}

	msg2 := nats.Msg{Data: []byte("ok")}
	return msg2
}

// createUser handles user creation requests.
func (uh *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Println("Error cant mimi.ParseMediaType")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		sendErrorWithMessage(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	log.Println("Pre decodeBody")
	rt, err := decodeBody(req.Body)
	if err != nil {
		if strings.Contains(err.Error(), "Key: 'User.Username' Error:Field validation for 'Username' failed on the 'min' tag") {
			sendErrorWithMessage(w, "Username must have minimum 6 characters", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Key: 'User.Password' Error:Field validation for 'Password' failed on the 'password' tag") {
			sendErrorWithMessage(w, "Password must have minimum 8 characters,minimum one big letter, numbers and special characters", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag") {
			sendErrorWithMessage(w, "Email format is incorrect", http.StatusBadRequest)
		} else {
			sendErrorWithMessage(w, "Ovde"+err.Error(), http.StatusBadRequest)
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
		sendErrorWithMessage(w, "", http.StatusInternalServerError)
	}
	rt.Password = hashedPassword
	log.Println("Hashed Password: %w", rt.Password)

	response, err := uh.db.Insert(rt)
	if err != nil {
		log.Println("Nije uspesno poslan user")
		if strings.Contains(err.Error(), "username") {
			sendErrorWithMessage(w, "Provide different username", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "email") {
			sendErrorWithMessage(w, "Provide differen email", http.StatusBadRequest)
		} else {
			sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		sendErrorWithMessage(w, "Error reading response body", http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()
	if string(responseBody) == "User created" {
		content := `
		// 		<h1>Verify your email</h1>
		// 		<h1>This is a verification message from AirBnb</h1>
		// 		<h4>Use the following code: %s</h4>
		// 		<h4><a href="https://localhost:4200/verify-email">Click here</a> to verify your email.</h4>`
		subject := "Verification email"
		uh.sendEmail(rt, content, subject, true, rt.Email)
		sendErrorWithMessage(w, "User cretated. Check the email for verification code", http.StatusCreated)
	} else {
		sendErrorWithMessage(w, string(responseBody), response.StatusCode)
	}

	// w.WriteHeader(http.StatusCreated)
	// if response != "" {
	// 	if strings.Contains(response, "username") {
	// 		sendErrorWithMessage(w, response, http.StatusBadRequest)
	// 	} else if strings.Contains(response, "Provide different email") {
	// 		sendErrorWithMessage(w, response, http.StatusBadRequest)
	// 	} else if strings.Contains(response, "User not created") {
	// 		sendErrorWithMessage(w, response, http.StatusBadRequest)
	// 	} else if response == "ok" {
	// 		err := uh.Create(rt.Username, rt.Password, rt.Email)
	// 		if err != nil {
	// 			sendErrorWithMessage(w, "User not succesfully created", http.StatusBadRequest)
	// 			return
	// 		}
	// 		content := `
	// 		<h1>Verify your email</h1>
	// 		<h1>This is a verification message from AirBnb</h1>
	// 		<h4>Use the following code: %s</h4>
	// 		<h4><a href="https://localhost:4200/verify-email">Click here</a> to verify your email.</h4>`
	// 		subject := "Verification email"
	// 		uh.sendEmail(rt, content, subject, true, rt.Email)

	// 		w.WriteHeader(http.StatusCreated)
	// 	}
	// 	return
	// }
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
		sendErrorWithMessage(w, "Authorization payload not found", http.StatusInternalServerError)
		return
	}

	// Check user role for authorization
	if authPayload.Role == "guest" {
		sendErrorWithMessage(w, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	// Convert users to JSON and send the response
	err = users.ToJSON(w)
	if err != nil {
		sendErrorWithMessage(w, "Unable to convert to JSON", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to JSON:", err)
		return
	}
}

// loginUser handles user login requests.
func (uh *UserHandler) loginUser(w http.ResponseWriter, req *http.Request) {
	rt, err := decodeLoginBody(req.Body)
	if err != nil {
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := rt.Username
	password := rt.Password
	user, err := uh.db.GetByUsername(username)
	if err != nil {
		log.Println("mongo: no documents in result: treba da se registuje neko")
		sendErrorWithMessage(w, "No such user", http.StatusBadRequest)
		return
	}

	// prooveravamo da li korisnik ima verifikovan mejl 169,170,171,172,173,174
	log.Println(user.IsEmailVerified)
	if !user.IsEmailVerified {
		sendErrorWithMessage(w, "Email is not verifyed (treba da postoji dugme da se ponovo posalje email ili da mu se napise da proveri email ako nije)", http.StatusBadRequest)
		return
	}

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	// If user is not found, return an error
	if user == nil {
		sendErrorWithMessage(w, "Invalid username or password", http.StatusNotFound)
		return
	}

	// Check if the provided password matches the hashed password in the database
	err = CheckHashedPassword(password, user.Password)
	if err != nil {
		sendErrorWithMessage(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate and send JWT token as a response
	jwtToken(user, w, uh)
}

func (uh *UserHandler) sendEmail(newUser *User, contentStr string, subjectStr string, isVerificationEmail bool, email string) error { // ako isVerificationEmial is true than VrificationEmail is sending and if is false ForgottenPasswordEmial is sending
	log.Println("SendEmail()")

	randomCode := randstr.String(20)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	log.Println(tlsConfig)

	err := uh.isVerificationEmail(newUser, randomCode, isVerificationEmail)
	if err != nil {
		return err
	}

	sender := mail.NewGmailSender("Air Bnb", "mobilneaplikcijesit@gmail.com", "esrqtcomedzeapdr", tlsConfig) //postavi recoveri password
	subject := subjectStr
	content := fmt.Sprintf(contentStr, randomCode)
	to := []string{email}
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

func (uh *UserHandler) sendForgottenPasswordEmail(w http.ResponseWriter, req *http.Request) {
	log.Println("Usli u sendForgottenPasswordEmail")
	vars := mux.Vars(req)
	email := vars["email"]

	allValidEmails, err := uh.db.GetAllVerificationEmailsByEmail(email) // provera ako neko probad a posalje mejl a nije registrovan
	if err != nil {
		sendErrorWithMessage(w, "Error in geting AllVerificationEmails"+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(email)
	log.Println(allValidEmails)
	if len(allValidEmails) == 0 {
		sendErrorWithMessage(w, "No valid verification emails found for the given email 1", http.StatusBadRequest)
		return
	}

	succes := false
	for _, ve := range allValidEmails { // prolazimo kroz sve emejlove koje smo dobili sa emejlom koji smo poslali da bi videli da li je mejl verifikovan ako jeste onda moze da se posalje mejl korisniku da za zaboravljenu sifru na mejl koji je poslao
		if ve.IsUsed {
			succes = true
			break
		}
	}

	if succes {
		log.Println("Usli u succes")
		content := `
				<h1>Reset Your Password</h1>
				<h1>This is a password reset message from AirBnb</h1>
				<h4>Code for password reset: %s</h4>`
		subject := "Password Reset"
		user := &User{
			Email: email,
		}
		err := uh.sendEmail(user, content, subject, false, email)
		if err != nil {
			sendErrorWithMessage(w, "Cant send email "+err.Error(), http.StatusBadRequest)
			return
		}

		sendErrorWithMessage(w, "Please check emial for verification code", http.StatusOK)
	} else {
		sendErrorWithMessage(w, "No valid verification emails found for the given email 2", http.StatusBadRequest)
		return
	}
}
func (uh *UserHandler) isVerificationEmail(newUser *User, randomCode string, isVerificationEmail bool) error {
	log.Println("Usli u isVerificationEmail")
	if isVerificationEmail {
		verificationEmail := VerifyEmail{
			Username:   newUser.Username,
			Email:      newUser.Email,
			SecretCode: randomCode,
			IsUsed:     false,
			CreatedAt:  time.Now(),
			ExpiredAt:  time.Now().Add(15 * time.Minute), // moras da promenis da je trajanje 15 min
		}

		log.Println("Verifikacioni mejl: ", verificationEmail)

		err := uh.db.CreateVerificationEmail(verificationEmail)
		if err != nil {
			log.Println("Cant save verification email in SendEmail()method")
			return err
		}
	} else {
		forgottenPasswordEmail := ForgottenPasswordEmail{
			Email:      newUser.Email,
			SecretCode: randomCode,
			IsUsed:     false,
			CreatedAt:  time.Now(),
			ExpiredAt:  time.Now().Add(15 * time.Minute), // moras da promenis da je trajanje 15 min
		}

		log.Println("ForgottenPassword mejl: ", forgottenPasswordEmail)
		err := uh.db.CreateForgottenPasswordEmail(forgottenPasswordEmail)
		if err != nil {
			log.Println("Cant save forgotten password email in SendEmail()method")
			return err
		}
	}
	return nil
}

func (uh *UserHandler) changeForgottenPassword(w http.ResponseWriter, req *http.Request) {
	rt, err := decodeForgottenPasswordBody(req.Body)
	if err != nil {
		if strings.Contains(err.Error(), "Key: 'ForgottenPassword.NewPassword' Error:Field validation for 'NewPassword' failed on the 'newPassword' tag") {
			sendErrorWithMessage(w, "NewPassword must have minimum 8 characters,minimum one big letter, one number and special characters", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "required") {
			sendErrorWithMessage(w, "All fealds are required", http.StatusBadRequest)
		} else {
			sendErrorWithMessage(w, "Ovde "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	if rt.NewPassword != rt.ConfirmPassword {
		sendErrorWithMessage(w, "Confirm password must be same as New password", http.StatusBadRequest)
		return
	}

	forgottenPasswordEmail, err := uh.db.GetForgottenPasswordEmailByCode(rt.Code)
	if err != nil {
		log.Println("Error in getting Email by code:", err)
		sendErrorWithMessage(w, "Code is not valid", http.StatusBadRequest)
		return
	}

	if forgottenPasswordEmail != nil {
		if !forgottenPasswordEmail.IsUsed {
			isActive, err := uh.db.IsForgottenPasswordEmailActive(rt.Code)
			if err != nil {
				log.Println("Error Code is not active")
				sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
				return
			}
			if isActive {

				// verifikacija passworda treba da se radi odmah u decodBodiu

				sanitizedPassword := sanitizeInput(rt.NewPassword)

				blacklist, err := NewBlacklistFromURL()
				if err != nil {
					log.Println("Error fetching blacklist: %v\n", err)
					return
				}

				if blacklist.IsBlacklisted(rt.NewPassword) {
					log.Println("Password is too weak, blacklist")
					sendErrorWithMessage(w, "Password is too weak", http.StatusBadRequest)
					return
				}

				user := &UserA{
					Username:        "",
					Password:        sanitizedPassword,
					Email:           forgottenPasswordEmail.Email,
					IsEmailVerified: true,
				}

				hashedPassword, err := HashPassword(sanitizedPassword)
				if err != nil {
					sendErrorWithMessage(w, "", http.StatusInternalServerError)
				}

				user.Password = hashedPassword

				err = uh.db.UpdateUsersPassword(user)
				if err != nil {
					log.Println("Error when updating password")
					sendErrorWithMessage(w, "Error when updating password "+err.Error(), http.StatusBadRequest)
					return
				}

				err = uh.db.UpdateForgottenPasswordEmail(rt.Code)
				if err != nil {
					log.Println("Error in trying to update VerificationEmail")
					sendErrorWithMessage(w, "Error in trying to update VerificationEmail", http.StatusInternalServerError)
					return
				}

				sendErrorWithMessage(w, "Password succesfuly changed", http.StatusOK)

			} else {
				sendErrorWithMessage(w, "Code is not active", http.StatusBadRequest)
				return
			}
		} else {
			sendErrorWithMessage(w, "Code that has been forwarded has been used", http.StatusBadRequest)
			return
		}
	}
}

func (uh *UserHandler) verifyEmail(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	code := vars["code"]

	verificationEmail, err := uh.db.GetVerificationEmailByCode(code)
	if err != nil {
		log.Println("Error in getting verificationEmail:", err)
		sendErrorWithMessage(w, "Error in getting verificationEmail", http.StatusInternalServerError)
		return
	}

	if verificationEmail != nil {
		if !verificationEmail.IsUsed {
			isActive, err := uh.db.IsVerificationEmailActive(code)
			if err != nil {
				log.Println("Error Verification code is not active")
				sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
				return
			}
			if isActive {
				err = uh.db.UpdateUsersVerificationEmail(verificationEmail.Username)
				if err != nil {
					log.Println("Error in trying to update UsersVerificationEmail")
					sendErrorWithMessage(w, "Error in trying to update UsersVerificationEmail", http.StatusInternalServerError)
					return
				}

				err = uh.db.UpdateVerificationEmail(code)
				if err != nil {
					log.Println("Error in trying to update VerificationEmail")
					sendErrorWithMessage(w, "Error in trying to update VerificationEmail", http.StatusInternalServerError)
					return
				}

				sendErrorWithMessage(w, "Your mail have been verified", http.StatusAccepted)
			} else {
				sendErrorWithMessage(w, "Code is not active", http.StatusBadRequest)
				return
			}
		} else {
			sendErrorWithMessage(w, "Code that has been forwarded has been used", http.StatusBadRequest)
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
		sendErrorWithMessage(w, err.Error(), http.StatusInternalServerError)
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

func decodeForgottenPasswordBody(r io.Reader) (*ForgottenPassword, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt ForgottenPassword
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	if err := ValidateForgottenPassword(rt); err != nil {
		log.Println("ForgottenPasswordCredentials are not succesfuly validated in ValidateForgottenPassword func")
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
		log.Println("Ovde")
		sendErrorWithMessage(w, err.Error(), http.StatusInternalServerError)
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

func sendErrorWithMessage(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"message": message}
	json.NewEncoder(w).Encode(errorResponse)
}
