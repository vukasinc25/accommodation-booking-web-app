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

	service := &accoHandler{
		db: map[string]*Accommodation{},
	}
	router.HandleFunc("/accommodation", service.createAccommodation).Methods("POST")
	router.HandleFunc("/accommodations", service.getAllAccommodations).Methods("GET")

	// start servergo get -u github.com/gorilla/mux

	srv := &http.Server{Addr: config["address"], Handler: router}
	go func() {
		log.Println("server starting")
		log.Println("Port:", config["address"])
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

func loadConfig() map[string]string {
	config := make(map[string]string)
	config["host"] = os.Getenv("HOST")
	config["port"] = os.Getenv("PORT")
	config["address"] = fmt.Sprintf(":%s", os.Getenv("PORT"))
	return config
}
