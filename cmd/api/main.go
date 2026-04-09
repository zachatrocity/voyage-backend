// @title Voyage API
// @version 1.0
// @description A self-hosted travel plan aggregator that searches through emails
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8181
// @BasePath /api/v1
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/zachatrocity/voyage/docs" // Import generated docs
	"github.com/zachatrocity/voyage/internal/api/handlers"
	voyagemiddleware "github.com/zachatrocity/voyage/internal/api/middleware"
	"github.com/zachatrocity/voyage/internal/trips"
)

// Config holds all configuration for the Voyage API server.
type Config struct {
	Port              string // VOYAGE_PORT (default "8181")
	AllowedOrigins    string // VOYAGE_ALLOWED_ORIGINS (comma-separated)
	MailDir           string // VOYAGE_MAIL_DIR
	NotmuchConfig     string // VOYAGE_NOTMUCH_CONFIG
	TripsDatabasePath string // VOYAGE_TRIPS_DB (default "./trips.json")
	ClassifiersPath   string // VOYAGE_CLASSIFIERS (default "./configs/classifiers.yaml")
	SyncCmd           string // VOYAGE_SYNC_CMD (default "./scripts/sync-mail.sh")
	APIKey            string // VOYAGE_API_KEY (empty = auth disabled)
}

func loadConfig() Config {
	return Config{
		Port:              getEnvOrDefault("VOYAGE_PORT", "8181"),
		AllowedOrigins:    os.Getenv("VOYAGE_ALLOWED_ORIGINS"),
		MailDir:           os.Getenv("VOYAGE_MAIL_DIR"),
		NotmuchConfig:     os.Getenv("VOYAGE_NOTMUCH_CONFIG"),
		TripsDatabasePath: getEnvOrDefault("VOYAGE_TRIPS_DB", "./trips.json"),
		ClassifiersPath:   getEnvOrDefault("VOYAGE_CLASSIFIERS", "./configs/classifiers.yaml"),
		SyncCmd:           getEnvOrDefault("VOYAGE_SYNC_CMD", "./scripts/sync-mail.sh"),
		APIKey:            os.Getenv("VOYAGE_API_KEY"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func main() {
	cfg := loadConfig()

	// Log loaded configuration
	apiKeyStatus := "disabled"
	if cfg.APIKey != "" {
		apiKeyStatus = "enabled (set)"
	}
	log.Printf("Voyage config: port=%s mail_dir=%s notmuch_config=%s trips_db=%s classifiers=%s sync_cmd=%s api_key=%s",
		cfg.Port, cfg.MailDir, cfg.NotmuchConfig, cfg.TripsDatabasePath, cfg.ClassifiersPath, cfg.SyncCmd, apiKeyStatus)

	// Initialize trips service
	tripStore, err := trips.NewJSONStore(cfg.TripsDatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize trip store: %v", err)
	}
	tripSvc := trips.NewService(tripStore)
	tripHandler := handlers.NewTripHandler(tripSvc)

	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS configuration
	// Default to wildcard for self-hosted deployments (mobile apps send null/app origin)
	origins := []string{"*"}
	if cfg.AllowedOrigins != "" {
		origins = strings.Split(cfg.AllowedOrigins, ",")
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept, "X-API-Key"},
	}))

	// Serve Swagger JSON file
	e.Static("/swagger", "./docs")

	// Scalar API documentation endpoint
	e.GET("/docs", func(c echo.Context) error {
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Voyage API Documentation",
			},
			DarkMode: true,
		})
		if err != nil {
			log.Printf("Error generating API documentation: %v", err)
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to generate API documentation: %v", err))
		}
		return c.HTML(http.StatusOK, htmlContent)
	})

	// API key auth
	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey != "" {
		log.Printf("API key authentication enabled")
	} else {
		log.Printf("API key authentication disabled (set VOYAGE_API_KEY to enable)")
	}

	// API v1 group
	v1 := e.Group("/api/v1")
	v1.Use(voyagemiddleware.APIKeyMiddleware(apiKey))
	{
		e.GET("/health", handlers.HealthCheck)

		// Search endpoint
		v1.GET("/search", handlers.Search)

		// Email endpoints
		v1.GET("/email/:id", handlers.GetEmail)
		v1.GET("/email/:id/content", handlers.GetEmailContent)

		// Tag email endpoint
		v1.POST("/email/:id/tags/:tag", handlers.TagEmail)

		// Trip endpoints
		v1.GET("/trips", tripHandler.ListTrips)
		v1.POST("/trips", tripHandler.CreateTrip)
		v1.DELETE("/trips/:id", tripHandler.DeleteTrip)
		v1.GET("/trips/:id/emails", tripHandler.ListTripEmails)

		// Email-trip association endpoint
		v1.POST("/email/:id/trip/:tripId", tripHandler.AssociateTripEmail)
	}

	// Start the server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
