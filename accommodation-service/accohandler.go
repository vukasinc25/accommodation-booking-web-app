package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type KeyProduct struct{}

type accoHandler struct {
	logger *log.Logger
	db     *AccoRepo
}

//type accoHandler struct {
//	db           map[string]*Accommodation // izigrava bazu podataka
//	accomodation *Accommodation
//}

func NewAccoHandler(l *log.Logger, r *AccoRepo) *accoHandler {
	return &accoHandler{l, r}
}

func (ah *accoHandler) createAccommodation(rw http.ResponseWriter, req *http.Request) {

	accommodation := req.Context().Value(KeyProduct{}).(*Accommodation)
	ah.db.Insert(accommodation)
	rw.WriteHeader(http.StatusCreated)

	//contentType := req.Header.Get("Content-Type")
	//mediatype, _, err := mime.ParseMediaType(contentType)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//if mediatype != "application/json" {
	//	err := errors.New("Expect application/json Content-Type")
	//	http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
	//	return
	//}
	//
	//rt, err := decodeBody(req.Body)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	////id := uuid.New().String()
	////ah.db[id] = rt
	//renderJSON(w, rt)
}

func (ah *accoHandler) getAllAccommodations(rw http.ResponseWriter, req *http.Request) {
	//newAccomodation := Accommodation{
	//	ID:   1,
	//	Name: "Sample Accommodation",
	//}
	//renderJSON(w, newAccomodation)
	// allAccommodations := []*Accommodation{}
	// for _, v := range ah.db {
	// 	allAccommodations = append(allAccommodations, v)
	// }
	accommodations, err := ah.db.GetAll()
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		return
	}

	err = accommodations.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (ah *accoHandler) MiddlewareAccommodationDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		accommodation := &Accommodation{}
		err := accommodation.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			ah.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, accommodation)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}

func (ah *accoHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		ah.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
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
