package hilo

import (
	"github.com/JILI-GAMES/b_backend_games/pkg/common/logger"
	"strings"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"
	"github.com/gofiber/fiber/v2"
)

// RouteGroup holds the dependencies for the handlers
type RouteGroup struct {
	GameLogger   *logger.GameLogger
	RNGProd      *rng.Client
	SettingsProd *settings.Client
	RNGTest      *rng.Client
	SettingsTest *settings.Client
}

// NewRouteGroup creates a new RouteGroup
func NewRouteGroup(rngProd *rng.Client, settingsProd *settings.Client, rngTest *rng.Client, settingsTest *settings.Client) *RouteGroup {
	return &RouteGroup{
		RNGProd:      rngProd,
		SettingsProd: settingsProd,
		RNGTest:      rngTest,
		SettingsTest: settingsTest,
	}
}

// Helper to select the correct clients per request
func (rg *RouteGroup) getClientsForRequest(c *fiber.Ctx) (*rng.Client, *settings.Client) {
	origin := c.Get("Origin")
	rg.GameLogger.Debug("Origin: %s", origin)
	if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test")) || origin == "" {
		return rg.RNGTest, rg.SettingsTest
	}
	return rg.RNGProd, rg.SettingsProd
}

// Register registers the routes with the Fiber app
func (rg *RouteGroup) Register(app *fiber.App) {
	// Pre-game endpoints (NEW)
	app.Post("/preview/hilo", rg.PreviewHandler)
	app.Post("/preview-skip/hilo", rg.PreviewSkipHandler)

	// Main game endpoints
	app.Post("/start/hilo", rg.StartGameHandler)
	app.Post("/guess/hilo", rg.GuessHandler)
	app.Post("/skip/hilo", rg.SkipHandler)
	app.Post("/cashout/hilo", rg.CashoutHandler)
	app.Post("/verify/hilo", rg.VerifyDeckHandler)

	// Optional utility endpoint for getting current betting options
	app.Post("/options/hilo", rg.GetOptionsHandler)
}

// Optional endpoint to get betting options for current card
func (rg *RouteGroup) GetOptionsHandler(c *fiber.Ctx) error {
	var req struct {
		CurrentCard string `json:"current_card"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request",
		})
	}

	if req.CurrentCard == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Current card is required",
		})
	}

	options := GetBaseHiloOptions(req.CurrentCard) // Base multipliers for utility endpoint

	return c.JSON(fiber.Map{
		"status":        "success",
		"current_card":  req.CurrentCard,
		"current_value": GetCardValue(req.CurrentCard),
		"bet_options":   options,
	})
}
