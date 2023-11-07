package main

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
)

type accoHandler struct {
	db map[string]*Accommodation // izigrava bazu podataka
}

func (ah *accoHandler) createAccommodation(w http.ResponseWriter, req *http.Request) {
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

	id := uuid.New().String()
	ah.db[id] = rt
	renderJSON(w, rt)
}

func (ah *accoHandler) getAllAccommodations(w http.ResponseWriter, req *http.Request) {
	allAccommodations := []*Accommodation{}
	for _, v := range ah.db {
		allAccommodations = append(allAccommodations, v)
	}

	renderJSON(w, allAccommodations)
}

func decodeBody(r io.Reader) (*Accommodation, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt Accommodation
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
