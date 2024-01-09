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

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sony/gobreaker"
)

func main() {

	config := loadConfig()

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

	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[accommo-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[accommo-store] ", log.LstdFlags)
	//pub := InitPubSub()
	store, err := New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	store.Ping()

	router := mux.NewRouter()
	//router.StrictSlash(true)
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	service := NewAccoHandler(logger, store)

	router.Use(service.MiddlewareContentTypeSet)

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/api/accommodations/create", service.createAccommodation)
	postRouter.Use(service.MiddlewareRoleCheck(authClient, authBreaker))
	postRouter.Use(service.MiddlewareAccommodationDeserialization)

	router.HandleFunc("/api/accommodations/", service.getAllAccommodations).Methods("GET")
	router.HandleFunc("/api/accommodations/{id}", service.GetAccommodationById).Methods("GET")
	router.HandleFunc("/api/accommodations/myAccommodations/{username}", service.GetAllAccommodationsByUsername).Methods("GET")
	router.HandleFunc("/api/accommodations/search_by_location/{locations}", service.GetAllAccommodationsByLocation).Methods("GET")
	router.HandleFunc("/api/accommodations/search_by_noGuests/{noGuests}", service.GetAllAccommodationsByNoGuests).Methods("GET")
	router.HandleFunc("/api/accommodations/get_all_acco_by_id/{id}", service.GetAllAccommodationsById).Methods("GET")

	server := http.Server{
		Addr:         ":" + config["port"],
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("Server listening on port", config["port"])
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	signal.Notify(sigCh, syscall.SIGKILL)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)
	timeoutContext, _ = context.WithTimeout(context.Background(), 30*time.Second)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")

}

func loadConfig() map[string]string {
	config := make(map[string]string)
	config["host"] = os.Getenv("HOST")
	config["port"] = os.Getenv("PORT")
	config["address"] = fmt.Sprintf(":%s", os.Getenv("PORT"))
	return config
}
