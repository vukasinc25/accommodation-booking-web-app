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

type reservationHandler struct {
	logger *log.Logger
	repo   *ReservationRepo
}

func NewReservationHandler(l *log.Logger, r *ReservationRepo) *reservationHandler {
	return &reservationHandler{l, r}
}

func (rh *reservationHandler) Test(res http.ResponseWriter, req *http.Request) {
	log.Println("AAA")
}
func (rh *reservationHandler) GetAllReservationIds(res http.ResponseWriter, req *http.Request) {
	reservationIds, err := rh.repo.GetDistinctIds("reservation_id", "reservations_by_user")
	if err != nil {
		rh.logger.Print("Database exception: ", err)
	}

	if reservationIds == nil {
		return
	}

	rh.logger.Println(reservationIds)

	e := json.NewEncoder(res)
	err = e.Encode(reservationIds)
	if err != nil {
		http.Error(res, "Unable to convert to json", http.StatusInternalServerError)
		rh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (rh *reservationHandler) GetReservationDatesByAccomodationId(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	accoId := vars["id"]

	reservationDatesByAccomodationId, err := rh.repo.GetReservationsDatesByAccomodationId(accoId)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		sendErrorWithMessage(res, "Error when getting reservation dates", http.StatusBadRequest)
		return
	}

	if reservationDatesByAccomodationId == nil {
		return
	}

	err = reservationDatesByAccomodationId.ToJSON(res)
	if err != nil {
		sendErrorWithMessage(res, "Unable to convert to json", http.StatusInternalServerError)
		rh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

// func (rh *reservationHandler) createReservation(w http.ResponseWriter, req *http.Request) {
// 	contentType := req.Header.Get("Content-Type")
// 	mediatype, _, err := mime.ParseMediaType(contentType)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	if mediatype != "application/json" {
// 		err := errors.New("Expect application/json Content-Type")
// 		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
// 		return
// 	}

// 	rt, err := decodeBody(req.Body)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	rh.repo.Insert(rt)
// 	w.WriteHeader(http.StatusCreated)
// }

func (rh *reservationHandler) GetAllReservationsByAccomodationId(res http.ResponseWriter, req *http.Request) {
	log.Println("Usli u GetAllReservationsByAccomodationId")
	vars := mux.Vars(req)
	accoId := vars["id"]

	reservationsByAcco, err := rh.repo.GetReservationsByAcco(accoId)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		sendErrorWithMessage(res, "Error when getting reservations", http.StatusBadRequest)
		return
	}

	if reservationsByAcco == nil {
		return
	}

	err = reservationsByAcco.ToJSON(res)
	if err != nil {
		sendErrorWithMessage(res, "Unable to convert to json", http.StatusInternalServerError)
		rh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (rh *reservationHandler) getAllReservationsByUser(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]

	reservationsByUser, err := rh.repo.GetReservationsByUser(userId)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
	}

	if reservationsByUser == nil {
		return
	}

	err = reservationsByUser.ToJSON(res)
	if err != nil {
		http.Error(res, "Unable to convert to json", http.StatusInternalServerError)
		rh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (rh *reservationHandler) CreateReservationDateForAccomodation(res http.ResponseWriter, req *http.Request) {
	log.Println("Usli u Metodu")
	reservationDate, err := decodeBody(req.Body)
	if err != nil {
		log.Println("Error in decoding body")
		sendErrorWithMessage(res, "Error in decoding body", http.StatusBadRequest)
		return
	}
	// reservationDateByAccomodation := req.Context().Value(KeyProduct{}).(*ReservationDateByAccomodationId)
	log.Println(reservationDate.AccoId)
	log.Println(reservationDate.BeginAccomodationDate)
	log.Println(reservationDate.EndAccomodationDate)

	err = rh.repo.InsertReservationDateForAccomodation(reservationDate)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		sendErrorWithMessage(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
}

func (rh *reservationHandler) CreateReservationForAcco(res http.ResponseWriter, req *http.Request) {
	log.Println("Usli u CreateReservationForAcco")
	reservation, err := decodeReservationBody(req.Body)
	if err != nil {
		log.Println("Error in decoding body")
		sendErrorWithMessage(res, "Error in decoding body", http.StatusBadRequest)
		return
	}
	err = rh.repo.InsertReservationByAcco(reservation)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		sendErrorWithMessage(res, "Cant create reservation", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
}

func (rh *reservationHandler) CreateReservationForUser(res http.ResponseWriter, req *http.Request) {
	reservationUser := req.Context().Value(KeyProduct{}).(*ReservationByUser)
	// provera da li je odredjeni period dostupnostio dostupan ili da ovde ne ide provera nego da se stavi provera kada se uzitava period dostupnosti za acomodaciju
	// lista koju bi napravili na osnovu perioda rezevacije
	// for petlja kroz period rezervacije i upis za svaki period u bazu
	err := rh.repo.InsertReservationByUser(reservationUser)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
}

func (rh *reservationHandler) UpdateReservationByAcco(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	accoId := vars["accoId"]
	reservationId := vars["reservationId"]
	price := vars["price"]

	var stepenStudija string
	d := json.NewDecoder(req.Body)
	d.Decode(&stepenStudija)

	err := rh.repo.UpdateReservationByAcco(accoId, reservationId, price)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
}

func (rh *reservationHandler) MiddlewareReservationForAccoDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		reservationByAcco := &ReservationByAccommodation{}
		err := reservationByAcco.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			rh.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, reservationByAcco)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (rh *reservationHandler) MiddlewareReservationForUserDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		reservationByUser := &ReservationByUser{}
		err := reservationByUser.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			rh.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, reservationByUser)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

// func (ah *reservationHandler) MiddlewareAccommodationDeserialization(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
// 		accommodation := &Accommodation{}
// 		err := accommodation.FromJSON(h.Body)
// 		if err != nil {
// 			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
// 			ah.logger.Fatal(err)
// 			return
// 		}

// 		ctx := context.WithValue(h.Context(), KeyProduct{}, accommodation)
// 		h = h.WithContext(ctx)

// 		next.ServeHTTP(rw, h)
// 	})
// }

func (rh *reservationHandler) MiddlewareRoleCheck(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
			rh.logger.Println("Host Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Println("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}
			rh.logger.Println(resp)

			hostID := string(resBody)

			var requestBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				rh.logger.Println("Error decoding request body:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if startDateStr, ok := requestBody["startDate"].(string); ok { //parsiranje datuma
				startDate, err := time.Parse("2006-01-02", startDateStr)
				if err != nil {
					rh.logger.Println("Error parsing startDate:", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				requestBody["startDate"] = startDate
			}

			if endDateStr, ok := requestBody["endDate"].(string); ok { //
				endDate, err := time.Parse("2006-01-02", endDateStr)
				if err != nil {
					rh.logger.Println("Error parsing endDate:", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				requestBody["endDate"] = endDate
			}

			hostID = strings.Trim(hostID, `"`)
			requestBody["hostId"] = hostID

			modifiedJSON, err := json.Marshal(requestBody)
			if err != nil {
				rh.logger.Println("Error marshaling modified JSON:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), bytes.NewBuffer(modifiedJSON))
			if err != nil {
				rh.logger.Println("Error creating new request:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newReq.Header = r.Header

			newReq.Header.Set("Content-Type", "application/json")
			log.Println(newReq.Body)

			next.ServeHTTP(w, newReq)
		})
	}
}

func (rh *reservationHandler) MiddlewareRoleCheck1(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
			rh.logger.Println("Host Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Println("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}
			rh.logger.Println(resp)

			next.ServeHTTP(w, r)
		})
	}
}

func (rh *reservationHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		rh.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

// func decodeBody(r io.Reader) (*Reservation, error) {
// 	dec := json.NewDecoder(r)
// 	dec.DisallowUnknownFields()

// 	var rt Reservation
// 	if err := dec.Decode(&rt); err != nil {
// 		return nil, err
// 	}
// 	return &rt, nil
// }

// func renderJSON(w http.ResponseWriter, v interface{}) {
// 	js, err := json.Marshal(v)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(js)
// }

//	func (u *Reservations) ToJSON(w io.Writer) error {
//		e := json.NewEncoder(w)
//		return e.Encode(u)
//	}
func decodeBody(r io.Reader) (*ReservationDateByAccomodationId, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt ReservationDateByAccomodationId
	if err := dec.Decode(&rt); err != nil {
		log.Println("Error u decode body:", err)
		return nil, err
	}

	return &rt, nil
}

func decodeReservationBody(r io.Reader) (*ReservationByAccommodation, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt ReservationByAccommodation
	if err := dec.Decode(&rt); err != nil {
		log.Println("Error u decode body:", err)
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
