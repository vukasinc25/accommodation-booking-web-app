package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sony/gobreaker"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type AccoHandler struct {
	logger *log.Logger
	db     *AccoRepo
}

func NewAccoHandler(l *log.Logger, r *AccoRepo) *AccoHandler {

	return &AccoHandler{l, r}
}

func (ah *AccoHandler) createAccommodation(rw http.ResponseWriter, req *http.Request) {

	accommodation := req.Context().Value(KeyProduct{}).(*Accommodation)
	ah.logger.Println(accommodation)
	err := ah.db.Insert(accommodation)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusCreated)

}

func (ah *AccoHandler) GetAccommodationById(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	accommodation, err := ah.db.GetById(id)
	if err != nil {
		ah.logger.Println(err)
	}

	if accommodation == nil {
		http.Error(w, "Accommodation with given id not found", http.StatusNotFound)
		ah.logger.Printf("Accommodation with id: '%s' not found", id)
		return
	}

	err = accommodation.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (ah *AccoHandler) GetAllAccommodationsById(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	accommodations, err := ah.db.GetAllById(id)
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Accommodations with given username not found", http.StatusNotFound)
		ah.logger.Printf("Accommodations with username: '%s' not found", id)
		return
	}

	err = accommodations.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
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

func (ah *AccoHandler) MiddlewareRoleCheck(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			reqURL := "http://auth-service:8000/api/users/auth"

			authorizationHeader := r.Header.Get("authorization")
			fields := strings.Fields(authorizationHeader)
			if len(fields) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			log.Println(fields)

			accessToken := fields[1]

			var token ReqToken
			token.Token = accessToken

			jsonToken, _ := json.Marshal(token)

			cbResp, err := breaker.Execute(func() (interface{}, error) {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, bytes.NewBuffer(jsonToken))
				if err != nil {
					return nil, err
				}
				return client.Do(req)
			})
			if err != nil {
				ah.logger.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			ah.logger.Println(string(resBody))
			if resp.StatusCode != http.StatusOK {
				ah.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				ah.logger.Println("status " + resp.Status)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ah.logger.Println(resp)

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
