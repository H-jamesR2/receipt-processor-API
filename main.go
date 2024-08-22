package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"rcpt-proc-challenge-ans/config"
	"rcpt-proc-challenge-ans/controller"
	"rcpt-proc-challenge-ans/middleware"

	"github.com/gorilla/mux"
	_ "rcpt-proc-challenge-ans/docs"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"github.com/pressly/goose/v3"
    _ "github.com/lib/pq" // PostgreSQL driver for Goose
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

	config.Init()
	defer config.Log.Sync()

	runMigrations()

	r := mux.NewRouter()
	//r.Use(middleware.PreProcessLoggingMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.StripSlash)

    // Use StrictSlash(true) to automatically redirect trailing slash requests
    r.StrictSlash(true)
	

	r.HandleFunc("/receipts/process", controller.ProcessReceipt).Methods("POST")
	r.HandleFunc("/receipts/{id}", controller.GetReceipt).Methods("GET")
	r.HandleFunc("/receipts/{id}/points", controller.GetReceiptPoints).Methods("GET")
	r.HandleFunc("/receipts", controller.GetAllReceipts).Methods("GET")
	
	// Handle all other routes
    r.NotFoundHandler = http.HandlerFunc(controller.NotFoundHandler)
    r.MethodNotAllowedHandler = http.HandlerFunc(controller.MethodNotAllowedHandler)


	// Swagger documentation endpoint
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	
	config.Log.Info("Server is running", zap.String("port", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func runMigrations() {
    dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
        os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

    db, err := goose.OpenDBWithDriver("postgres", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    if err := goose.Up(db, "db/migrations"); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
}