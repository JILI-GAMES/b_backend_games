package opensesame1

import (
	"fmt"
	"strings"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"
	"github.com/gofiber/fiber/v2"
)

// RouteGroup holds the dependencies for the handlers
type RouteGroup struct {
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
	fmt.Printf("Origin: %s\n", origin)
	if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test")) {
		return rg.RNGTest, rg.SettingsTest
	}
	return rg.RNGProd, rg.SettingsProd
}

// Register registers the routes with the Fiber app
func (rg *RouteGroup) Register(app *fiber.App) {
	app.Post("/spin/opensesame1", rg.SpinHandler)
	app.Post("/select-free-spin-option/opensesame1", rg.SelectFreeSpinOptionHandler)
}
