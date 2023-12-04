package main

import (
	"Rest/data"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "[reservation-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[patient-store] ", log.LstdFlags)

	store, err := data.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseSession()
	store.CreateTables()

	router := mux.NewRouter()
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
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
