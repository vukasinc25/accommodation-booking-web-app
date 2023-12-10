package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

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

func (rh *reservationHandler) getAllReservationsByAcco(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	accoId := vars["id"]

	reservationsByAcco, err := rh.repo.GetReservationsByAcco(accoId)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
	}

	if reservationsByAcco == nil {
		return
	}

	err = reservationsByAcco.ToJSON(res)
	if err != nil {
		http.Error(res, "Unable to convert to json", http.StatusInternalServerError)
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

func (rh *reservationHandler) CreateReservationForAcco(res http.ResponseWriter, req *http.Request) {
	reservationAcco := req.Context().Value(KeyProduct{}).(*ReservationByAccommodation)
	err := rh.repo.InsertReservationByAcco(reservationAcco)
	if err != nil {
		rh.logger.Print("Database exception: ", err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
}

func (rh *reservationHandler) CreateReservationForUser(res http.ResponseWriter, req *http.Request) {
	reservationUser := req.Context().Value(KeyProduct{}).(*ReservationByUser)
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

// func (u *Reservations) ToJSON(w io.Writer) error {
// 	e := json.NewEncoder(w)
// 	return e.Encode(u)
// }
