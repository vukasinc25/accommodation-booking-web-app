package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	// "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vukasinc25/fst-airbnb/cache"
	"github.com/vukasinc25/fst-airbnb/handlers"
	"github.com/vukasinc25/fst-airbnb/storage"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	lumberjack "github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {

	logger := log.New()

	// Set up log rotation with Lumberjack
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "/acoo/file.log",
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

	//TRACING
	tracerProvider, err := NewTracerProvider(config["jaeger"])
	if err != nil {
		log.Fatal("JaegerTraceProvider failed to Initialize", err)
	}
	tracer := tracerProvider.Tracer("accommodation-service")
	//

	timeoutContext, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// logger := log.New(os.Stdout, "[accommo-api] ", log.LstdFlags)
	// storeLogger := log.New(os.Stdout, "[accommo-store] ", log.LstdFlags)
	// storageLogger := log.New(os.Stdout, "[file-storage] ", log.LstdFlags)
	// loggerCache := log.New(os.Stdout, "[redis-cache] ", log.LstdFlags)
	//pub := InitPubSub()
	store, err := New(timeoutContext, logger, config["conn_reservation_service_address"], tracer)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	store.Ping()

	// NoSQL: Initialize File Storage store
	// imageStore, err := storage.New(storageLogger)
	imageStore, err := storage.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	// Close connection to HDFS on shutdown
	defer func() {
		if err := imageStore.Close(); err != nil {
			log.Println("Error closing image store:", err)
		}
	}()

	// Create directory tree on HDFS
	_ = imageStore.CreateDirectories()

	// prCache := cache.New(loggerCache)
	prCache := cache.New(logger)
	// Test connection
	prCache.Ping()

	//Initialize the handler and inject said logger
	storageHandler := handlers.NewStorageHandler(logger, imageStore, prCache)

	router := mux.NewRouter()
	//router.StrictSlash(true)
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	service := NewAccoHandler(logger, store, storageHandler, tracer)

	router.Use(service.MiddlewareContentTypeSet)
	router.Use(service.ExtractTraceInfoMiddleware)

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/api/accommodations/create", service.createAccommodation)
	postRouter.Use(service.MiddlewareRoleCheck(authClient, authBreaker))
	postRouter.Use(service.MiddlewareAccommodationDeserialization)
	postRouter.Use(service.ExtractTraceInfoMiddleware)

	router.HandleFunc("/api/accommodations/", service.getAllAccommodations).Methods("GET")
	router.HandleFunc("/api/accommodations/{id}", service.GetAccommodationById).Methods("GET")
	router.HandleFunc("/api/accommodations/myAccommodations/{username}", service.GetAllAccommodationsByUsername).Methods("GET")
	router.HandleFunc("/api/accommodations/search_by_location/{locations}", service.GetAllAccommodationsByLocation).Methods("GET")
	router.HandleFunc("/api/accommodations/search_by_noGuests/{noGuests}", service.GetAllAccommodationsByNoGuests).Methods("GET")
	router.HandleFunc("/api/accommodations/get_all_acco_by_id/{id}", service.GetAllAccommodationsById).Methods("GET")
	router.HandleFunc("/api/accommodations/delete/{username}", service.DeleteAccommodation).Methods("DELETE")
	createAccommodationGrade := router.Methods(http.MethodPost).Subrouter()
	createAccommodationGrade.HandleFunc("/api/accommodations/accommodationGrade", service.GradeAccommodation) // treba authorisation
	createAccommodationGrade.Use(service.MiddlewareRoleCheck00(authClient, authBreaker))
	// router.HandleFunc("/api/accommodations/accommodationGrade", service.GradeAccommodation).Methods("POST")
	getAllAccommodationGrades := router.Methods(http.MethodGet).Subrouter()
	getAllAccommodationGrades.HandleFunc("/api/accommodations/accommodationGrades/{id}", service.GetAllAccommodationGrades)
	getAllAccommodationGrades.Use(service.MiddlewareRoleCheck(authClient, authBreaker))
	deleteAccommodationGrade := router.Methods(http.MethodDelete).Subrouter()
	deleteAccommodationGrade.HandleFunc("/api/accommodations/deleteAccommodationGrade/{id}", service.DeleteAccommodationGrade)
	deleteAccommodationGrade.Use(service.MiddlewareRoleCheck00(authClient, authBreaker))

	router.HandleFunc("/api/accommodations/copy", storageHandler.CopyFileToStorage).Methods("POST")

	router.HandleFunc("/api/accommodations/write", storageHandler.WriteFileToStorage).Methods("POST")

	getAccommodationImage := router.Methods(http.MethodGet).Subrouter()
	getAccommodationImage.HandleFunc("/api/accommodations/read/{fileName}", storageHandler.ReadFileFromStorage)
	getAccommodationImage.Use(storageHandler.MiddlewareCacheHit)

	server := http.Server{
		Addr:         ":" + config["port"],
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
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
	config["jaeger"] = os.Getenv("JAEGER_ADDRESS")
	config["conn_reservation_service_address"] = fmt.Sprintf("http://%s:%s", os.Getenv("RESERVATION_SERVICE_HOST"), os.Getenv("RESERVATION_SERVICE_PORT"))
	return config
}

func NewTracerProvider(collectorEndpoint string) (*sdktrace.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(collectorEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize exporter due: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("accommodation-service"),
			semconv.DeploymentEnvironmentKey.String("development"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
