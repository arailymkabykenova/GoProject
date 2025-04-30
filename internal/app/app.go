package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/cors"
	httpHandlers "template/internal/deliveries/http"
	"template/internal/repositories"
	"template/internal/services"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	log.Println("Starting application setup...")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/shortener.db"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	listenAddr := ":" + serverPort

	log.Printf("Database Path: %s", dbPath)
	log.Printf("Base URL: %s", baseURL)
	log.Printf("Server Port: %s", serverPort)

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory '%s': %v", dbDir, err)
	}
	db, err := repositories.ConnectDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}()

	log.Println("Initializing dependencies...")
	shortenerRepo := repositories.NewSQLiteShortenerRepo(db)
	if err := shortenerRepo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	shortenerService := services.NewShortenerService(shortenerRepo)
	shortenerHandler := httpHandlers.NewShortenerHandler(shortenerService, shortenerRepo, baseURL)

	log.Println("Setting up HTTP router...")
	mux := http.NewServeMux()
	shortenerHandler.RegisterRoutes(mux)

	log.Println("Configuring CORS...")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"null", "http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})
	handler := c.Handler(mux)

	log.Printf("Starting HTTP server on %s", listenAddr)
	server := &http.Server{
		Addr:         listenAddr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Println("Server stopped gracefully.")
	return nil
}
