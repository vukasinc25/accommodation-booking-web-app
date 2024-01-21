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
	"github.com/thanhpk/randstr"
	"github.com/vukasinc25/fst-airbnb/handlers"
)

type KeyProduct struct{}

type AccoHandler struct {
	logger         *log.Logger
	db             *AccoRepo
	storageHandler *handlers.StorageHandler
}

func NewAccoHandler(l *log.Logger, r *AccoRepo, sh *handlers.StorageHandler) *AccoHandler {

	return &AccoHandler{l, r, sh}
}

func (ah *AccoHandler) createAccommodation(rw http.ResponseWriter, req *http.Request) {

	accommodation := req.Context().Value(KeyProduct{}).(*Accommodation)
	ah.logger.Println(accommodation)
	accommodation.AverageGrade = 0
	err := ah.db.Insert(accommodation)
	if err != nil {
		log.Println("error:1", err.Error())
		if strings.Contains(err.Error(), "duplicate key") {
			sendErrorWithMessage(rw, "accommodation with that name already exists", http.StatusBadRequest)
			return
		}
		sendErrorWithMessage(rw, "", http.StatusBadRequest)
		return
	}
	// ah.storageHandler.WriteFileToStorage(rw, req)
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

func (ah *AccoHandler) GetAllAccommodationsByUsername(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	accommodations, err := ah.db.GetAllByUsername(username)
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Accommodations with given username not found", http.StatusNotFound)
		ah.logger.Printf("Accommodations with username: '%s' not found", username)
		return
	}

	err = accommodations.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (ah *AccoHandler) GetAllAccommodationsById(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	log.Println(id)
	accommodations, err := ah.db.GetAllById(id)
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Accommodations with given id not found", http.StatusNotFound)
		ah.logger.Printf("Accommodations with id: '%s' not found", id)
		return
	}

	err = accommodations.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (ah *AccoHandler) GetAllAccommodationsByLocation(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	locations := vars["locations"]
	accommodations, err := ah.db.GetAllByLocation(locations)
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Accommodations with given location not found", http.StatusNotFound)
		ah.logger.Printf("Accommodations with location: '%s' not found", locations)
		return
	}

	err = accommodations.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		ah.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (ah *AccoHandler) DeleteAccommodationGrade(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	userId, ok := req.Context().Value("userId").(string)
	if !ok {
		log.Println("Error retrieving userId from context")
		sendErrorWithMessage1(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Println("UserId:", userId)
	log.Println("AccommodationGradeId:", id)

	err := ah.db.DeleteAccommodationGrade(userId, id)
	if err != nil {
		log.Println("Error2:", err)
		sendErrorWithMessage(res, err.Error(), http.StatusInternalServerError)
		return
	}

	sendErrorWithMessage(res, "Accommodation grade succesfully deleted", http.StatusOK)
}

func (ah *AccoHandler) DeleteAccommodation(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	err := ah.db.Delete(username)
	if err != nil {
		log.Println("Error when tried to delete accommodation:", err)
		sendErrorWithMessage1(res, "Error when tried to delete accommodation", http.StatusInternalServerError)
		return
	}
	sendErrorWithMessage1(res, "User succesfully deleted", http.StatusOK)
}

func (rh *AccoHandler) MiddlewareRoleCheck00(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			reqURL := "http://auth-service:8000/api/users/auth"

			authorizationHeader := r.Header.Get("authorization")
			fields := strings.Fields(authorizationHeader)

			if len(fields) == 0 {
				sendErrorWithMessage(w, "", http.StatusUnauthorized)
				return
			}

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
				rh.logger.Println(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			rh.logger.Println("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Println("status " + resp.Status)
				sendErrorWithMessage(w, "Lavor", resp.StatusCode)
				return
			}

			userID := string(resBody)
			ctx = context.WithValue(ctx, "userId", userID)
			ctx = context.WithValue(ctx, "accessToken", accessToken)

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), r.Body)
			if err != nil {
				rh.logger.Println("Error creating new request:", err)
				sendErrorWithMessage(w, "Error creating new request:"+err.Error(), http.StatusInternalServerError)
				return
			}
			newReq.Header = r.Header

			newReq.Header.Set("Content-Type", "application/json")

			log.Println("Token:", token)
			log.Println("AccessToken:", accessToken)

			next.ServeHTTP(w, newReq)
		})
	}
}

func (ah *AccoHandler) GetAllAccommodationGrades(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	accommodationId := vars["id"]

	accommodationGrades, err := ah.db.GetAllAccommodationGrades(accommodationId)
	if err != nil {
		log.Println("Error1:", err.Error())
		sendErrorWithMessage(res, "Error in GetAllAccommodationGrades method", http.StatusInternalServerError)
		return
	}

	err = accommodationGrades.ToJSON(res)
	if err != nil {
		log.Println("Unable to convert to json :", err)
		sendErrorWithMessage(res, "Unable to convert to json", http.StatusInternalServerError)
		return
	}
}

func (ah *AccoHandler) GetAllAccommodationsByNoGuests(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	noGuests := vars["noGuests"]

	accommodations, err := ah.db.GetAllByNoGuests(noGuests)
	if err != nil {
		ah.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Accommodations with given noGuests not found", http.StatusNotFound)
		ah.logger.Printf("Accommodations with noGuests: '%s' not found", noGuests)
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
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			ah.logger.Println(string(resBody))
			if resp.StatusCode != http.StatusOK {
				ah.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				ah.logger.Println("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}
			ah.logger.Println(resp)

			next.ServeHTTP(w, r)
		})
	}
}

func (ah *AccoHandler) GradeAccommodation(res http.ResponseWriter, req *http.Request) {
	log.Println("Request Body:", req.Body)
	accommodationGrade, err := decodeAccommodatioGradeBody(req.Body)
	if err != nil {
		log.Println("Cant decode body")
		sendErrorWithMessage(res, err.Error(), http.StatusBadRequest)
		return
	}

	userId, ok := req.Context().Value("userId").(string)
	if !ok {
		log.Println("Error retrieving userId from context")
		sendErrorWithMessage1(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	accommodationGrade.CreatedAt = formattedTime
	accommodationGrade.ID = randstr.String(20)
	accommodationGrade.UserId = strings.Trim(userId, `"`)
	log.Println("HostGrade:", accommodationGrade)

	tocken, ok := req.Context().Value("accessToken").(string)
	if !ok {
		log.Println("Error retrieving tocken from context")
		sendErrorWithMessage1(res, "", http.StatusInternalServerError)
		return
	}

	err = ah.db.CreateGrade(accommodationGrade, tocken)
	if err != nil {
		log.Println("Error in inserting accommodation grade")
		sendErrorWithMessage(res, err.Error(), http.StatusInternalServerError)
		return
	}

	sendErrorWithMessage(res, "Accommodation succesfuly created", http.StatusCreated)
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

func decodeAccommodatioGradeBody(r io.Reader) (*AccommodationGrade, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt AccommodationGrade
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	err := ValidateAccommodationGrade(&rt)
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

func sendErrorWithMessage(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"message": message}
	json.NewEncoder(w).Encode(errorResponse)
}

func sendErrorWithMessage1(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
	w.WriteHeader(statusCode)
}
