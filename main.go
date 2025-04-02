// main.go
package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    
    "github.com/JILI-GAMES/b_backend_games/internal/config"
    "github.com/JILI-GAMES/b_backend_games/internal/platform/rng"
    "github.com/JILI-GAMES/b_backend_games/internal/platform/settings"
	"github.com/JILI-GAMES/b_backend_games/games/superace_deluxe"
    // Future games will be imported here
)

func main() {
    cfg := config.Load()
    
    // Initialize service clients
    rngClient := rng.NewClient(cfg.RNGServiceURL)
    settingsClient := settings.NewClient(cfg.SettingsServiceURL)
    
    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
    })
    
    app.Use(loggerMiddleware)
    
    superaceRoutes := superace_deluxe.NewRouteGroup(rngClient, settingsClient)
    superaceRoutes.Register(app)
    
    // Future games will register their routes here
    
    // Start the server
    log.Printf("Server starting on :%s", cfg.ServerPort)
    if err := app.Listen(":" + cfg.ServerPort); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func customErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    return c.Status(code).JSON(fiber.Map{
        "status":  code,
        "message": err.Error(),
    })
}

func loggerMiddleware(c *fiber.Ctx) error {
    log.Printf("[%s] %s", c.Method(), c.Path())
    return c.Next()
}