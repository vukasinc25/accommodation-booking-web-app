package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/vukasinc25/fst-airbnb/token"
)

type UserHandler struct {
	logger   *log.Logger
	db       *UserRepo
	jwtMaker token.Maker
}

func NewUserHandler(l *log.Logger, r *UserRepo, jwtMaker token.Maker) *UserHandler {
	return &UserHandler{l, r, jwtMaker}
}

func (uh *UserHandler) Auth(w http.ResponseWriter, req *http.Request) {

	header := req.Header.Get("Authorization")

	uh.logger.Println(header)

	if req.Header.Get("X-Original-Uri") == "/api/accommodations/" {
		w.WriteHeader(http.StatusOK)
		return
	}

	fields := strings.Fields(header)
	_, err := uh.jwtMaker.VerifyToken(fields[1])

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

}

func (uh *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := decodeBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blacklist, err := NewBlacklistFromURL()
	if err != nil {
		log.Println("Error fetching blacklist: %v\n", err)
		return
	}

	if blacklist.IsBlacklisted(rt.Password) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Password nije dobar")
		return
	}

	log.Println("Not hashed Password: %w", rt.Password)
	hashedPassword, err := HashPassword(rt.Password)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	rt.Password = hashedPassword
	log.Println("Hashed Password: %w", rt.Password)

	uh.db.Insert(rt)
	w.WriteHeader(http.StatusCreated)
}

func (uh *UserHandler) getAllUsers(w http.ResponseWriter, req *http.Request) {
	users, err := uh.db.GetAll()
	ctx := req.Context()

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	if users == nil {
		return
	}

	authPayload, ok := ctx.Value(AuthorizationPayloadKey).(*token.Payload)
	if !ok || authPayload == nil {
		http.Error(w, "Authorisation payload not found", http.StatusInternalServerError)
		return
	}

	log.Println("Authorisation payload:", authPayload)

	err = users.ToJSON(w)
	if err != nil {
		http.Error(w, "Unable to convert to json", http.StatusInternalServerError)
		uh.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (uh *UserHandler) loginUser(w http.ResponseWriter, req *http.Request) {
	rt, err := decodeLoginBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := rt.Username
	password := rt.Password

	user, err := uh.db.GetByUsername(username)

	if err != nil {
		uh.logger.Print("Database exception: ", err)
	}

	if user == nil {
		http.Error(w, "invalid username or password", http.StatusNotFound)
		return
	}

	err = CheckHashedPassword(password, user.Password)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		uh.logger.Println("token encoding error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jwtToken(user, w, uh)
}

func jwtToken(user *User, w http.ResponseWriter, uh *UserHandler) {
	durationStr := "15m" // treba da bude u konstanta izvan funkcije
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Println("Cant make duration")
		return
	}

	accessToken, accessPayload, err := uh.jwtMaker.CreateToken(
		user.ID,
		user.Username,
		user.Role,
		duration,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp := LoginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	e := json.NewEncoder(w)
	e.Encode(rsp)
}

func decodeBody(r io.Reader) (*User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt User
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func decodeLoginBody(r io.Reader) (*LoginUser, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt LoginUser
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

func (u *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}
