// @title Voyage API
// @version 1.0
// @description A self-hosted travel plan aggregator that searches through emails
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/zachatrocity/voyage/docs" // Import generated docs
	"github.com/zachatrocity/voyage/internal/api/handlers"
)

func main() {
	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/health", handlers.HealthCheck)

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

	// API v1 group
	v1 := e.Group("/api/v1")
	{
		// Search endpoint
		v1.GET("/search", handlers.Search)

		// Email endpoint
		v1.GET("/email/:id", handlers.GetEmail)

		// Tag email endpoint
		v1.POST("/email/:id/tags/:tag", handlers.TagEmail)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
