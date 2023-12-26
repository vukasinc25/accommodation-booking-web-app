package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

type userHandler struct {
	logger *log.Logger
	db     *UserRepo
}

func NewUserHandler(l *log.Logger, r *UserRepo) *userHandler {
	return &userHandler{l, r}
}

func (rh *userHandler) MiddlewareRoleCheck(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
			rh.logger.Println("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Println("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Println("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}

			userID := string(resBody)

			requestBody := map[string]interface{}{
				"userId": userID,
			}

			userID = strings.Trim(userID, `"`)
			requestBody["userId"] = userID

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

			next.ServeHTTP(w, newReq)
		})
	}
}

func (uh *userHandler) createUser(w http.ResponseWriter, req *http.Request) {
	log.Println("Usli u CreateUser")
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Println("Error cant mimi.ParseMediaType")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		sendErrorWithMessage(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := decodeBody(req.Body)
	if err != nil {
		log.Println("Cant decode user")
		sendErrorWithMessage(w, "Cant decode user", http.StatusNotAcceptable)
		return
	}

	err = uh.db.Insert(rt)
	if err != nil {
		log.Println("User not saved")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendErrorWithMessage(w, "User created", http.StatusCreated)
}

func (uh *userHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {

	// log.Println("Get All Users method enterd geting Accomodation")
	// uh.getAccomodations(w)
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
		log.Println("Lavor", r)
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
func (u *ResponseUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (uh *userHandler) GetUserById(res http.ResponseWriter, req *http.Request) {

	requestId, err := decodeIdBody(req.Body)
	if err != nil {
		log.Println("Cant decode body")
		sendErrorWithMessage(res, "Cant decode body", http.StatusBadRequest)
		return
	}

	log.Println("usao u metodu")

	user, err := uh.db.Get(requestId.UserId)
	if err != nil {
		http.Error(res, "Database exception", http.StatusInternalServerError)
		uh.logger.Fatal("Database exception: ", err)
	}

	if user == nil {
		http.Error(res, "User with given email not found", http.StatusNotFound)
		uh.logger.Printf("User with email: '%s' not found", &requestId)
		return
	}

	err = user.ToJSON(res)
	if err != nil {
		http.Error(res, "Unable to convert to json", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func decodeIdBody(r io.Reader) (*RequestId, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt RequestId
	if err := dec.Decode(&rt); err != nil {
		log.Println("Error u decode body:", err)
		return nil, err
	}

	return &rt, nil
}

func sendErrorWithMessage(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
	w.WriteHeader(statusCode)
}

func sendErrorWithMessage1(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"message": message}
	json.NewEncoder(w).Encode(errorResponse)
}
