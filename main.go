package main

import (
	"log"
	"os"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/config"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"

	"github.com/JILI-GAMES/b_backend_games/pkg/games/kong"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/superace_deluxe"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/magicace"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/moneybagsman2"


	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set up logging
	logFile, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Create shared clients
	rngClient := rng.NewClient(cfg.RNGServiceURL)
	settingsClient := settings.NewClient(cfg.SettingsServiceURL)

	// Create fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add middleware
	app.Use(recover.New())

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", 
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           86400, 
	}))

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     logFile,
	}))

	// Register routes for individual games
	superaceRoutes := superace_deluxe.NewRouteGroup(rngClient, settingsClient)
	superaceRoutes.Register(app)

	kongRoutes := kong.NewRouteGroup(rngClient, settingsClient)
	kongRoutes.Register(app)

	magicAceRoutes := magicace.NewRouteGroup(rngClient, settingsClient)
	magicAceRoutes.Register(app)

	moneyBagsMan2Routes := moneybagsman2.NewRouteGroup(rngClient, settingsClient)
	moneyBagsMan2Routes.Register(app)


	// Add a simple status endpoint
	app.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"games": []string{
				"superace_deluxe",
				"kong",
				"magicAce",
				"moneyBagsMan2",
			},
		})
	})

	// Start the server
	port := cfg.ServerPort
	log.Printf("Starting server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// Custom error handler
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	// Handle common errors with appropriate status codes
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("Error: %v", err)

	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
	})
}
