package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vukasinc25/fst-airbnb/token"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8000"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	router := mux.NewRouter()
	//router.StrictSlash(true)
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[auth-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[auth-store] ", log.LstdFlags)

	tokenMaker, err := token.NewJWTMaker("12345678901234567890123456789012") // 12345678901234567890123456789012 treba da bude izvan primajuceg parametra
	if err != nil {
		logger.Fatal(err)
	}

	// NoSQL: Initialize Product Repository store
	store, err := New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	// NoSQL: Checking if the connection was established
	store.Ping()

	service := NewUserHandler(logger, store, tokenMaker)
	authRoutes := router.PathPrefix("/").Subrouter()
	authRoutes.Use(AuthMiddleware(tokenMaker))

	router.HandleFunc("/api/users/auth", service.Auth)
	router.HandleFunc("/api/users/register", service.createUser).Methods("POST")
	router.HandleFunc("/api/users/login", service.loginUser).Methods("POST")
	authRoutes.HandleFunc("/api/users/users", service.getAllUsers).Methods("GET")

	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12, // or tls.VersionTLS13
			CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
			// Add other cipher suites as needed
		},
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		// err := server.ListenAndServe()
		err := server.ListenAndServeTLS("cert/auth-server.crt", "cert/auth-server.key")

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
