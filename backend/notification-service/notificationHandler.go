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

type NotificationHandler struct {
	logger *log.Logger
	db     *NotificationRepo
}

func NewNotificationHandler(l *log.Logger, r *NotificationRepo) *NotificationHandler {

	return &NotificationHandler{l, r}
}

func (nh *NotificationHandler) createNotification(rw http.ResponseWriter, req *http.Request) {
	notification, err := decodeBody(req.Body)
	notification.Date = time.Now()
	//notification := req.Context().Value(KeyProduct{}).(*Notification)
	err = nh.db.Insert(notification)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusCreated)

}

func (nh *NotificationHandler) GetNotificationById(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	notification, err := nh.db.GetById(id)
	if err != nil {
		nh.logger.Println(err)
	}

	if notification == nil {
		http.Error(w, "Notification with given id not found", http.StatusNotFound)
		nh.logger.Printf("Notification with id: '%s' not found", id)
		return
	}

	err = notification.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		nh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

//func (nh *NotificationHandler) GetAllNotificationsByUsername(w http.ResponseWriter, req *http.Request) {
//	vars := mux.Vars(req)
//	username := vars["username"]
//
//	accommodations, err := nh.db.GetAllByUsername(username)
//	if err != nil {
//		nh.logger.Print("Database exception: ", err)
//	}
//
//	if accommodations == nil {
//		http.Error(w, "Notifications with given username not found", http.StatusNotFound)
//		nh.logger.Printf("Notifications with username: '%s' not found", username)
//		return
//	}
//
//	err = accommodations.ToJSON(w)
//	if err != nil {
//		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
//		nh.logger.Fatal("Unable to convert to json :", err)
//		return
//	}
//}

func (nh *NotificationHandler) GetAllNotificationsByUserId(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	log.Println(id)
	accommodations, err := nh.db.GetAllByHostId(id)
	if err != nil {
		nh.logger.Print("Database exception: ", err)
	}

	if accommodations == nil {
		http.Error(w, "Notifications with given id not found", http.StatusNotFound)
		nh.logger.Printf("Notifications with id: '%s' not found", id)
		return
	}

	err = accommodations.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		nh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (nh *NotificationHandler) DeleteNotification(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	err := nh.db.Delete(username)
	if err != nil {
		log.Println("Error when tried to delete notification:", err)
		sendErrorWithMessage1(res, "Error when tried to delete notification", http.StatusInternalServerError)
		return
	}
	sendErrorWithMessage1(res, "User succesfully deleted", http.StatusOK)
}

func (nh *NotificationHandler) MiddlewareRoleCheck00(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
				nh.logger.Println(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			nh.logger.Println("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				nh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				nh.logger.Println("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}

			userID := string(resBody)
			ctx = context.WithValue(ctx, "userId", userID)
			ctx = context.WithValue(ctx, "accessToken", accessToken)

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), r.Body)
			if err != nil {
				nh.logger.Println("Error creating new request:", err)
				w.WriteHeader(http.StatusInternalServerError)
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

func (nh *NotificationHandler) MiddlewareNotificationDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		notification := &Notification{}
		err := notification.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			nh.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, notification)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}

func (nh *NotificationHandler) MiddlewareRoleCheck(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
				nh.logger.Println(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			nh.logger.Println(string(resBody))
			if resp.StatusCode != http.StatusOK {
				nh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				nh.logger.Println("status " + resp.Status)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			nh.logger.Println(resp)

			next.ServeHTTP(w, r)
		})
	}
}

func (nh *NotificationHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		nh.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func decodeBody(r io.Reader) (*Notification, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt Notification
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
