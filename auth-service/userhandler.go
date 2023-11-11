package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
)

type UserHandler struct {
	logger *log.Logger
	db     *UserRepo
}

func NewUserHandler(l *log.Logger, r *UserRepo) *UserHandler {
	return &UserHandler{l, r}
}

func (uh *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := decodeBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("Not hashed Password: %w", rt.Password)
	hashedPassword, err := HashPassword(rt.Password)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	rt.Password = hashedPassword
	log.Println("Hashed Password: %w", rt.Password)

	uh.db.Insert(rt)
	w.WriteHeader(http.StatusCreated)
}

func (uh *UserHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {
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

func (uh *UserHandler) getUserByUsername(w http.ResponseWriter, req *http.Request) {

	rt, err := decodeLoginBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := rt.Username
	password := rt.Password

	user, err := uh.db.GetByUsername(username)

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = CheckHashedPassword(password, user.Password)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// needs to return JWT
	err = user.ToJSON(w)
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
		return nil, err
	}
	return &rt, nil
}

func decodeLoginBody(r io.Reader) (*LoginUser, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt LoginUser
	if err := dec.Decode(&rt); err != nil {
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
