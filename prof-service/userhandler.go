package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
)

type userHandler struct {
	logger               *log.Logger
	db                   *UserRepo
	accomodation_address string
}

func NewUserHandler(l *log.Logger, r *UserRepo, accomodations_adrress string) *userHandler {
	return &userHandler{l, r, accomodations_adrress}
}

func (uh *userHandler) createUser(w http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != "application/json" {
		err := errors.New("Expect application/json Content-Type")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	rt, err := decodeBody(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uh.db.Insert(rt)
	w.WriteHeader(http.StatusCreated)
}

// type Accomodation struct {
// 	ID   int    `json:"ID"`
// 	Name string `json:"Name"`
// }

// func (uh *userHandler) getAccomodations(w http.ResponseWriter) (*Accomodation, error) {
// 	log.Println("Enterd in GetAccomodations")
// 	url := uh.accomodation_address + "/accommodations"
// 	log.Println("Accomodation address:", uh.accomodation_address)
// 	log.Println("Url:", url)
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		log.Println("Error in creating request")
// 		return nil, err
// 	}

// 	httpResp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println("Error in sendding or receving Resoponse from Accommodation service", req)
// 		return nil, err
// 	}
// 	log.Println("HttpRespons: ", httpResp)

// 	var resp *Accomodation

// 	err = json.NewDecoder(httpResp.Body).Decode(&resp)
// 	if err != nil {
// 		log.Println("Error decoding response")
// 		return nil, err
// 	}

//		renderJSON(w, resp)
//		return resp, nil
//	}
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
