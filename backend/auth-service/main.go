package main

import (
	"context"
	"fmt"

	// "log"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	lumberjack "github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"github.com/vukasinc25/fst-airbnb/token"
)

func main() {

	logger := log.New()

	// Set up log rotation with Lumberjack
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "/auth/file.log",
		MaxSize:    10, // MB
		MaxBackups: 3,
		LocalTime:  true, // Use local time
	}
	logger.SetOutput(lumberjackLogger)

	// Handle log rotation gracefully on program exit
	defer func() {
		if err := lumberjackLogger.Close(); err != nil {
			log.Error("Error closing log file:", err)
		}
	}()

	// ... (rest of your code)

	// Example log statements
	logger.Info("lavor1")

	config := loadConfig()
	// Read the port from the environment variable, default to "8000" if not set
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	// Create a context with a timeout of 50 seconds
	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// Initialize Gorilla Mux router and CORS middleware
	router := mux.NewRouter()
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	// Initialize loggers with prefixes for different components
	// logger := log.New(os.Stdout, "[auth-api] ", log.LstdFlags)
	// storeLogger := log.New(os.Stdout, "[auth-store] ", log.LstdFlags)

	// Create a JWT token maker
	tokenMaker, err := token.NewJWTMaker("12345678901234567890123456789012")
	if err != nil {
		logger.Fatal(err)
	}

	// NoSQL: Initialize auth Repository store
	store, err := New(timeoutContext, logger, config["conn_service_address"], config["conn_reservation_service_address"], config["conn_accommodation_service_address"])
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	// Check if the data store connection was established
	store.Ping()

	// Create a user handler service
	service := NewUserHandler(logger, store, tokenMaker)
	// subu := InitPubSubUsername()
	if err != nil {
		logger.Fatal(err)
	}
	// err = subu.Subscribe(func(msg *nats.Msg) {
	// 	pub, _ := nats2.NewNATSPublisher(msg.Reply)

	// 	response := service.SubscribeUsername(msg)

	// 	response.Reply = msg.Reply

	// 	pub.Publish(response)
	// })
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	authRoutes := router.PathPrefix("/").Subrouter()
	authRoutes.Use(AuthMiddleware(tokenMaker))

	router.HandleFunc("/api/users/auth", service.Auth).Methods("GET")
	router.HandleFunc("/api/users/register", SetCSPHeader(service.createUser)).Methods("POST") // uradjeno
	router.HandleFunc("/api/users/login", SetCSPHeader(service.loginUser)).Methods("POST")
	router.HandleFunc("/api/users/email/{code}", service.verifyEmail).Methods("POST")                            //uradjeno                         // for sending verification mail
	router.HandleFunc("/api/users/sendforgottemail/{email}", service.sendForgottenPasswordEmail).Methods("POST") // nije  // for sending forgotten password email
	router.HandleFunc("/api/users/changeForgottenPassword", service.changeForgottenPassword).Methods("POST")     // nije      // treba da se prosledi body sa newPassword, confirmPassword, code
	authRoutes.HandleFunc("/api/users/update", service.UpdateUser).Methods("PATCH")
	authRoutes.HandleFunc("/api/users/changePassword", service.ChangePassword).Methods("PATCH")
	authRoutes.HandleFunc("/api/users/delete", service.DeleteUser).Methods("DELETE")

	// Configure the HTTP server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// TLSConfig: &tls.Config{
		// 	InsecureSkipVerify: true, // samo za testiranje
		// MinVersion:         tls.VersionTLS12,
		// CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
		// },
	}

	// Print a message indicating the server is listening
	logger.Println("Server listening on port", port)

	// Start the HTTP server in a goroutine
	go func() {
		err := server.ListenAndServe()
		// err := server.ListenAndServeTLS("/cert/auth-server.crt", "/cert/auth-server.key")
		if err != nil {
			logger.Fatal(err)
		}
	}()

	// Listen for signals to gracefully shut down the server
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	signal.Notify(sigCh, syscall.SIGKILL)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	// Create a new context for graceful shutdown with a timeout of 30 seconds
	timeoutContext, _ = context.WithTimeout(context.Background(), 30*time.Second)

	// Attempt to gracefully shut down the server
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}

func loadConfig() map[string]string {
	config := make(map[string]string)
	config["conn_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("PROF_SERVICE_HOST"), os.Getenv("PROF_SERVICE_PORT"))
	config["conn_reservation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("RESERVATION_SERVICE_HOST"), os.Getenv("RESERVATION_SERVICE_PORT"))
	config["conn_accommodation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("ACCOMMODATION_SERVICE_HOST"), os.Getenv("ACCOMMODATION_SERVICE_PORT"))
	return config
}

// func logFile() {
// 	logFilePath := "/app/logs/all-services.log" // Use an absolute path inside the container

// 	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// 	if err != nil {
// 		log.Fatal("Error opening log file:", err)
// 	}
// 	defer logFile.Close()

// 	log.SetOutput(logFile)

// 	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
// }
