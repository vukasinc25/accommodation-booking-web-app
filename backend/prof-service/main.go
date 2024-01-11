package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

func main() {

	authClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
		},
	}

	authBreaker := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "auth",
			MaxRequests: 1,
			Timeout:     10 * time.Second,
			Interval:    0,
		})

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	router := mux.NewRouter()
	router.StrictSlash(true)

	config := loadConfig()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := New(logger, config["conn_reservation_service_address"])
	if err != nil {
		logger.Fatal(err)
	}

	service := NewUserHandler(logger, store)

	// router.HandleFunc("/api/prof/email/{code}", service.verifyEmail).Methods("POST") // for sending verification mail
	router.HandleFunc("/api/prof/create", service.createUser).Methods("POST")
	router.HandleFunc("/api/prof/users/", service.getAllUsers).Methods("GET")
	getUserInfoByUserId := router.Methods(http.MethodGet).Subrouter()
	getUserInfoByUserId.HandleFunc("/api/prof/user", service.GetUserById)
	getUserInfoByUserId.Use(service.MiddlewareRoleCheck(authClient, authBreaker))
	router.Methods(http.MethodPatch).Subrouter()
	getAllHostGrades := router.Methods(http.MethodGet).Subrouter()
	getAllHostGrades.HandleFunc("/api/prof/hostGrades/{id}", service.GetAllHostGrades) // treba authorisation
	// getAllHostGrades.Use(service.MiddlewareRoleCheck00(authClient, authBreaker))
	createHostGrade := router.Methods(http.MethodPost).Subrouter()
	createHostGrade.HandleFunc("/api/prof/hostGrade", service.CreateHostGrade) // treba authorisation
	createHostGrade.Use(service.MiddlewareRoleCheck0(authClient, authBreaker))
	deleteHostGrade := router.Methods(http.MethodDelete).Subrouter()
	deleteHostGrade.HandleFunc("/api/prof/hostGrade/{id}", service.DeleteHostGrade) // treba authorisation
	deleteHostGrade.Use(service.MiddlewareRoleCheck00(authClient, authBreaker))
	router.HandleFunc("/api/prof/update", service.UpdateUser).Methods("PATCH")
	router.HandleFunc("/api/prof/delete/{id}", service.DeleteUser).Methods("DELETE")

	// srv := &http.Server{Addr: config["address"], Handler: router}
	server := http.Server{
		Addr:         ":" + "8000",
		Handler:      router,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	go func() {
		log.Println("server starting")
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	<-quit

	log.Println("service shutting down ...")

	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")

}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func loadConfig() map[string]string {
	config := make(map[string]string)
	config["host"] = os.Getenv("HOST")
	config["port"] = os.Getenv("PORT")
	config["address"] = fmt.Sprintf(":%s", os.Getenv("PORT"))
	config["mondo_db_uri"] = os.Getenv("MONGO_DB_URI")
	config["conn_reservation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("RESERVATION_SERVICE_HOST"), os.Getenv("RESERVATION_SERVICE_PORT"))
	return config
}
