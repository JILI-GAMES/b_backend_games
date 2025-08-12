package main

import (
	"log"
	"os"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/config"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"

	"github.com/JILI-GAMES/b_backend_games/pkg/games/blossomsofwealth"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/kong"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/magicace"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/moneybagsman"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/moneybagsman2"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/opensesame1"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/opensesame2"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/superace_deluxe"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/winningmask"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/birdsparty"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/magicaceoriginal"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/hilo"
	"github.com/JILI-GAMES/b_backend_games/pkg/games/crazykingkong"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load both production and test configs
	prodCfg, testCfg := config.LoadAll()

	// Set up logging with lumberjack for daily rotation and 1 day retention (use prod config for log file)
	log.SetOutput(&lumberjack.Logger{
		Filename:  prodCfg.LogFile,
		MaxAge:    1,    // days to keep
		LocalTime: true, // use local time for file names
	})
	// Set up logging with lumberjack for daily rotation and 1 day retention (use test config for log file)
	log.SetOutput(&lumberjack.Logger{
		Filename:  testCfg.LogFile,
		MaxAge:    1,    // days to keep
		LocalTime: true, // use local time for file names
	})

	// Create both prod and test clients
	rngClientProd := rng.NewClient(prodCfg.RNGServiceURL)
	settingsClientProd := settings.NewClient(prodCfg.SettingsServiceURL)

	rngClientTest := rng.NewClient(testCfg.RNGServiceURL)
	settingsClientTest := settings.NewClient(testCfg.SettingsServiceURL)

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
		AllowHeaders:     "*",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     os.Stdout,
	}))

	// Register routes for individual games, passing both sets of clients
	superaceRoutes := superace_deluxe.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	superaceRoutes.Register(app)

	kongRoutes := kong.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	kongRoutes.Register(app)

	magicAceRoutes := magicace.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	magicAceRoutes.Register(app)

	moneyBagsMan2Routes := moneybagsman2.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	moneyBagsMan2Routes.Register(app)

	moneyBagsManRoutes := moneybagsman.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	moneyBagsManRoutes.Register(app)

	openSesame1Routes := opensesame1.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	openSesame1Routes.Register(app)

	blossomsofwealthRoutes := blossomsofwealth.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	blossomsofwealthRoutes.Register(app)

	openSesame2Routes := opensesame2.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	openSesame2Routes.Register(app)

	winningmaskRoutes := winningmask.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	winningmaskRoutes.Register(app)

	birdspartyRoutes := birdsparty.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	birdspartyRoutes.Register(app)
	
	magicAceOriginalRoutes := magicaceoriginal.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	magicAceOriginalRoutes.Register(app)

	hiloRoutes := hilo.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	hiloRoutes.Register(app)

	crazyKingKongRoutes := crazykingkong.NewRouteGroup(rngClientProd, settingsClientProd, rngClientTest, settingsClientTest)
	crazyKingKongRoutes.Register(app)

	// Add a simple status endpoint
	app.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"games": []string{
				"superace_deluxe",
				"kong",
				"magicAce",
				"moneyBagsMan2",
				"moneyBagsMan",
				"openSesame1",
				"blossomsofwealth",
				"openSesame2",
				"winningmask",
				"birdsparty",
				"magicAceOriginal",
				"hilo",
				"crazykingkong",
			},
		})
	})

	// Start the server
	port := prodCfg.ServerPort
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
