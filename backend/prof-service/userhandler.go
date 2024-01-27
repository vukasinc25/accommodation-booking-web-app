package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/thanhpk/randstr"
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
				rh.logger.Info(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			rh.logger.Info("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Info("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Info("status " + resp.Status)
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
				rh.logger.Info("Error marshaling modified JSON:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), bytes.NewBuffer(modifiedJSON))
			if err != nil {
				rh.logger.Info("Error creating new request:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newReq.Header = r.Header

			newReq.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, newReq)
		})
	}
}

func (rh *userHandler) MiddlewareRoleCheck0(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
				rh.logger.Info(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			rh.logger.Info("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Info("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Info("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}

			userID := string(resBody)

			requestBody := map[string]interface{}{}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				rh.logger.Info("Error decoding request body:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			userID = strings.Trim(userID, `"`)
			requestBody["userId"] = userID

			modifiedJSON, err := json.Marshal(requestBody)
			if err != nil {
				rh.logger.Info("Error marshaling modified JSON:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), bytes.NewBuffer(modifiedJSON))
			if err != nil {
				rh.logger.Info("Error creating new request:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newReq.Header = r.Header

			newReq.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, newReq)
		})
	}
}

func (rh *userHandler) MiddlewareRoleCheck00(client *http.Client, breaker *gobreaker.CircuitBreaker) mux.MiddlewareFunc {
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
				rh.logger.Info(err)
				sendErrorWithMessage(w, "Service is not working", http.StatusInternalServerError)
				return
			}

			resp := cbResp.(*http.Response)
			resBody, err := io.ReadAll(resp.Body)
			rh.logger.Info("User Id:", string(resBody))
			if resp.StatusCode != http.StatusOK {
				rh.logger.Info("Error in auth response " + strconv.Itoa(resp.StatusCode))
				rh.logger.Info("status " + resp.Status)
				w.WriteHeader(resp.StatusCode)
				return
			}

			userID := string(resBody)
			ctx = context.WithValue(ctx, "userId", userID)

			newReq, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), nil)
			if err != nil {
				rh.logger.Info("Error creating new request:", err)
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
	uh.logger.Info("Usli u CreateUser")
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		uh.logger.Info("Error cant mimi.ParseMediaType")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		sendErrorWithMessage(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	uh.logger.Info("Request User:", req.Body)
	rt, err := decodeBody(req.Body)
	if err != nil {
		uh.logger.Info("Cant decode user")
		sendErrorWithMessage(w, "Cant decode user", http.StatusNotAcceptable)
		return
	}

	err = uh.db.Insert(rt)
	if err != nil {
		uh.logger.Info("User not saved")
		sendErrorWithMessage(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendErrorWithMessage(w, "User created", http.StatusCreated)
}

func (uh *userHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {

	// uh.logger.Info("Get All Users method enterd geting Accomodation")
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

func (uh *userHandler) DeleteUser(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	err := uh.db.Delete(id)
	if err != nil {
		uh.logger.Info("Unable to delete product.", err)
		sendErrorWithMessage(res, err.Error(), http.StatusBadRequest)
		return
	}

	sendErrorWithMessage(res, "User succesfully deleted", http.StatusOK)
}

func decodeBody(r io.Reader) (*User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt User
	if err := dec.Decode(&rt); err != nil {
		log.Info("Lavor", err)
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

func (uh *userHandler) CreateHostGrade(res http.ResponseWriter, req *http.Request) {
	uh.logger.Info(req.Body)
	hostGrade, err := decodeHostGradeBody(req.Body)
	if err != nil {
		uh.logger.Info("Cant decode body")
		sendErrorWithMessage1(res, err.Error(), http.StatusBadRequest)
		return
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	hostGrade.CreatedAt = formattedTime
	hostGrade.ID = randstr.String(20)
	uh.logger.Info("HostGrade:", hostGrade)

	response, err := uh.db.GetAllReservatinsForUserByHostId(hostGrade.UserId, hostGrade.HostId)
	if err != nil {
		uh.logger.Info("Error in method GetAllReservatinsForUserByHostId", err)
		sendErrorWithMessage1(res, "Error in getting reservations for user", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		uh.logger.Info("Error in reading response body")
		sendErrorWithMessage1(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(string(body), "There is no active reservations for accommodations of this host") {
		sendErrorWithMessage1(res, "There is no active reservations for accommodations of this host", http.StatusBadRequest)
		return
	} else if strings.Contains(string(body), "There is no reservations for hosts accommodations") {
		sendErrorWithMessage1(res, "There is no reservations for hosts accommodations", http.StatusBadRequest)
		return
	} else if strings.Contains(string(body), "There is some reservtions for this user") {
		err = uh.db.CreateHostGrade(hostGrade)
		if err != nil {
			uh.logger.Info("HostGrade lavor")
			sendErrorWithMessage1(res, "Lavor when tryed to save HostGrade", http.StatusBadRequest)
			return
		}

		sendErrorWithMessage1(res, "Host grade is succesfully created", http.StatusOK)
		return
	} else if strings.Contains(string(body), "There is not reservations for hosts accommodations") {
		sendErrorWithMessage1(res, "There is not reservations for hosts accommodations", http.StatusBadRequest)
		return
	} else {
		sendErrorWithMessage1(res, string(body), http.StatusOK)
		return
	}
}

func (uh *userHandler) DeleteHostGrade(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	userId, ok := req.Context().Value("userId").(string)
	if !ok {
		uh.logger.Info("Error retrieving hostId from context")
		sendErrorWithMessage1(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err := uh.db.DeleteHostGrade(id, userId)
	if err != nil {
		uh.logger.Print("DeleteHost lavor")
		sendErrorWithMessage1(res, err.Error(), http.StatusBadRequest)
		return
	}

	sendErrorWithMessage1(res, "Host grade succesfully deleted", http.StatusOK)
}

func (uh *userHandler) GetAllHostGrades(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	// hostId, ok := req.Context().Value("hostId").(string)
	// if !ok {
	// 	uh.logger.Info("Error retrieving hostId from context")
	// 	sendErrorWithMessage1(res, "Internal Server Error", http.StatusInternalServerError)
	// 	return
	// }

	// uh.logger.Info("HostId", hostId)
	// uh.logger.Info("HostId")

	uh.logger.Info("Usli u GetAllHostGrades")
	hostGrades, err := uh.db.GetAllHostGradesByHostId(id)
	if err != nil {
		uh.logger.Info("GetAllHostGrades lavor")
		sendErrorWithMessage1(res, err.Error(), http.StatusBadRequest)
		return
	}

	if hostGrades == nil {
		sendErrorWithMessage1(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	e := json.NewEncoder(res)
	e.Encode(hostGrades)

}
func decodeUserInfoBody(r io.Reader) (*User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt User
	if err := dec.Decode(&rt); err != nil {
		log.Info("Lavor", r)
		return nil, err
	}

	if err := ValidateUser(&rt); err != nil {
		log.Info(err)
		return nil, err
	}
	return &rt, nil
}

func decodeHostGradeBody(r io.Reader) (*HostGrade, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt HostGrade
	if err := dec.Decode(&rt); err != nil {
		log.Info("Lavor", r)
		return nil, err
	}

	if err := ValidateHostGrade(&rt); err != nil {
		log.Info(err)
		return nil, err
	}
	return &rt, nil
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
func (uh *userHandler) UpdateUser(res http.ResponseWriter, req *http.Request) {
	uh.logger.Info("Usli u Update")
	uh.logger.Info(req.Body)

	user, err := decodeUserInfoBody(req.Body)
	if err != nil {
		uh.logger.Info("Cant decode body")
		sendErrorWithMessage1(res, "Cant decode body", http.StatusBadRequest)
		return
	}

	userDb, err := uh.db.Get(user.ID)
	if err != nil {
		uh.logger.Fatal("Database exception:", err)
		http.Error(res, "Database exception", http.StatusInternalServerError)
		return
	}

	if userDb == nil {
		uh.logger.Printf("Product with id: '%s' not found", user.ID)
		sendErrorWithMessage1(res, "Product with given id not found", http.StatusNotFound)
		return
	}

	user.Role = userDb.Role

	err = uh.db.UpdateUser(user)
	if err != nil {
		uh.logger.Info("Error in updating user: ", err)
		sendErrorWithMessage1(res, "Cant update user", http.StatusInternalServerError)
		return
	}

	sendErrorWithMessage(res, "User updated", http.StatusOK)
}

func (uh *userHandler) GetUserById(res http.ResponseWriter, req *http.Request) {

	requestId, err := decodeIdBody(req.Body)
	if err != nil {
		uh.logger.Info("Cant decode body")
		sendErrorWithMessage(res, "Cant decode body", http.StatusBadRequest)
		return
	}

	uh.logger.Info("usao u metodu")

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
		log.Info("Error u decode body:", err)
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
