package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
)

type userHandler struct {
	logger *log.Logger
	db     *UserRepo
}

func NewUserHandler(l *log.Logger, r *UserRepo) *userHandler {
	return &userHandler{l, r}
}

func (uh *userHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := decodeBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uh.db.Insert(rt)
	w.WriteHeader(http.StatusCreated)
}

func (uh *userHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {
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