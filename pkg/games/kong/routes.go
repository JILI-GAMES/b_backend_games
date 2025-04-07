package kong

import (
	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"
	"github.com/gofiber/fiber/v2"
)

// RouteGroup holds the dependencies for game routes
type RouteGroup struct {
	RNG      *rng.Client
	Settings *settings.Client
}

// NewRouteGroup creates a new route group for kong game
func NewRouteGroup(rngClient *rng.Client, settingsClient *settings.Client) *RouteGroup {
	return &RouteGroup{
		RNG:      rngClient,
		Settings: settingsClient,
	}
}

// Register registers the kong game routes
func (rg *RouteGroup) Register(app *fiber.App) {
	app.Post("/spin/kong", rg.SpinHandler)
}
