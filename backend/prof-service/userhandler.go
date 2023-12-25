package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
)

type userHandler struct {
	logger *log.Logger
	db     *UserRepo
}

func NewUserHandler(l *log.Logger, r *UserRepo) *userHandler {
	return &userHandler{l, r}
}

func (uh *userHandler) createUser(w http.ResponseWriter, req *http.Request) {
	log.Println("Usli u CreateUser")
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

	rt, err := decodeBody(req.Body)
	if err != nil {
		log.Println("Cant decode user")
		sendErrorWithMessage(w, "Cant decode user", http.StatusNotAcceptable)
		return
	}

	err = uh.db.Insert(rt)
	if err != nil {
		log.Println("User not saved")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendErrorWithMessage(w, "User created", http.StatusCreated)
}

func (uh *userHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {

	// log.Println("Get All Users method enterd geting Accomodation")
	// uh.getAccomodations(w)
	users, err := uh.db.GetAll()

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	if users == nil {
		return
	}

	err = users.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func decodeBody(r io.Reader) (*User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt User
	if err := dec.Decode(&rt); err != nil {
		log.Println("Lavor", r)
		return nil, err
	}
	return &rt, nil
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (u *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}
func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func sendErrorWithMessage(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
	w.WriteHeader(statusCode)
}
func (uh *userHandler) GetUserById(rw http.ResponseWriter, h *http.Request) {
	
	vars := mux.Vars(h)
	email := vars["email"]

	log.Println("usao u metodu")
	
	user, err := uh.db.Get(email)
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		uh.logger.Fatal("Database exception: ", err)
	}

	if user == nil {
		http.Error(rw, "User with given email not found", http.StatusNotFound)
		uh.logger.Printf("User with email: '%s' not found", email)
		return
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}