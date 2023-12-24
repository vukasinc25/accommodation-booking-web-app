package main

import (
	"context"
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

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[reservation-store] ", log.LstdFlags)

	store, err := New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables()

	reservationHandler := NewReservationHandler(logger, store)
	router := mux.NewRouter()
	router.HandleFunc("/api/reservations/r", reservationHandler.Test).Methods("GET")
	// router.Use(reservationHandler.MiddlewareContentTypeSet)

	getReservationIds := router.Methods(http.MethodGet).Subrouter()
	getReservationIds.HandleFunc("/api/reservations/", reservationHandler.GetAllReservationIds)

	// getReservationIds2 := router.Methods(http.MethodGet).Subrouter()
	// getReservationIds2.HandleFunc("/api/r", reservationHandler.GetAllReservationIds)

	getReservationsByAcco := router.Methods(http.MethodGet).Subrouter()
	getReservationsByAcco.HandleFunc("/api/reservations/by_acco/{id}", reservationHandler.GetAllReservationsByAccomodationId)
	getReservationsByAcco.Use(reservationHandler.MiddlewareRoleCheck1(authClient, authBreaker))

	getReservationsByUser := router.Methods(http.MethodGet).Subrouter()
	getReservationsByUser.HandleFunc("/api/reservations/by_acco", reservationHandler.getAllReservationsByUser)

	postReservationForAcco := router.Methods(http.MethodPost).Subrouter()
	postReservationForAcco.HandleFunc("/api/reservations/for_user", reservationHandler.CreateReservationForAcco)
	postReservationForAcco.Use(reservationHandler.MiddlewareReservationForAccoDeserialization)

	postReservationForUser := router.Methods(http.MethodPost).Subrouter()
	postReservationForUser.HandleFunc("/api/reservations/for_acco", reservationHandler.CreateReservationForAcco)
	postReservationForUser.Use(reservationHandler.MiddlewareRoleCheck(authClient, authBreaker))

	postReservationDateByAccomodation := router.Methods(http.MethodPost).Subrouter()
	postReservationDateByAccomodation.HandleFunc("/api/reservations/date_for_acoo", reservationHandler.CreateReservationDateForAccomodation)
	postReservationDateByAccomodation.Use(reservationHandler.MiddlewareRoleCheck1(authClient, authBreaker))

	getReservationDatesByAccomodationId := router.Methods(http.MethodGet).Subrouter()
	getReservationDatesByAccomodationId.HandleFunc("/api/reservations/dates_by_acco_id/{id}", reservationHandler.GetReservationDatesByAccomodationId)
	getReservationDatesByAccomodationId.Use(reservationHandler.MiddlewareRoleCheck1(authClient, authBreaker))

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	logger.Println("Server listening on port", port)
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