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
)

func main() {

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	router := mux.NewRouter()
	router.StrictSlash(true)

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	service := NewUserHandler(logger, store)

	// router.HandleFunc("/api/prof/email/{code}", service.verifyEmail).Methods("POST") // for sending verification mail
	router.HandleFunc("/api/prof/create", service.createUser).Methods("POST")
	router.HandleFunc("/api/prof/users/", service.getAllUsers).Methods("GET")
	router.HandleFunc("/api/prof/user/{email}", service.GetUserById).Methods("GET")

	// start servergo get -u github.com/gorilla/mux

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
	// config["accomodation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("ACCOMMODATIONS_SERVICE_HOST"), os.Getenv("ACCOMMODATIONS_SERVICE_PORT"))
	return config
}
