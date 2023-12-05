package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	utility "github.com/vukasinc25/fst-airbnb/utility/messaging"
	nats2 "github.com/vukasinc25/fst-airbnb/utility/messaging/nats"
)

type KeyProduct struct{}

type AccoHandler struct {
	logger *log.Logger
	db     *AccoRepo
}

func NewAccoHandler(l *log.Logger, r *AccoRepo) *AccoHandler {

	return &AccoHandler{l, r}
}

func InitPubSub() utility.Publisher {

	publisher, err := nats2.NewNATSPublisher("auth.check")
	if err != nil {
		log.Fatal(err)
	}
	return publisher
}

func (ah *AccoHandler) createAccommodation(rw http.ResponseWriter, req *http.Request) {

	accommodation := req.Context().Value(KeyProduct{}).(*Accommodation)
	err := ah.db.Insert(accommodation)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusCreated)

}

func (ah *AccoHandler) getAllAccommodations(rw http.ResponseWriter, req *http.Request) {

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

func (ah *AccoHandler) MiddlewareAccommodationDeserialization(next http.Handler) http.Handler {
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

func (ah *AccoHandler) MiddlewareRoleCheck(publisher utility.Publisher) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ah.logger.Println("Published")

			msg := nats.Msg{Data: []byte(r.Header.Get("Authorization"))}
			response, err := publisher.Publish(msg)
			if err != nil {
				return
			}

			ah.logger.Println(string(response.Data))
			if string(response.Data) != "ok" {
				w.WriteHeader(http.StatusUnauthorized)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (ah *AccoHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
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
