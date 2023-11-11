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

	config := loadConfig()

	router := mux.NewRouter()
	router.StrictSlash(true)

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[patient-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store
	store, err := New(timeoutContext, storeLogger, config["mondo_db_uri"])
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	service := NewUserHandler(logger, store, config["accomodation_service_address"])
	router.HandleFunc("/user/", service.createUser).Methods("POST")
	router.HandleFunc("/users/", service.getAllUsers).Methods("GET")

	// start servergo get -u github.com/gorilla/mux

	srv := &http.Server{Addr: config["address"], Handler: router}
	go func() {
		log.Println("server starting")
		if err := srv.ListenAndServe(); err != nil {
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

	if err := srv.Shutdown(ctx); err != nil {
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
	config["accomodation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("ACCOMMODATIONS_SERVICE_HOST"), os.Getenv("ACCOMMODATIONS_SERVICE_PORT"))
	return config
}
