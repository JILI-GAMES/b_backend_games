package superace_deluxe

import (
	"github.com/JILI-GAMES/b_backend_games/internal/platform/rng"
	"github.com/JILI-GAMES/b_backend_games/internal/platform/settings"

	"github.com/gofiber/fiber/v2"
)

type RouteGroup struct {
	RNG      *rng.Client
	Settings *settings.Client
}

func NewRouteGroup(rngClient *rng.Client, settingsClient *settings.Client) *RouteGroup {
	return &RouteGroup{
		RNG:      rngClient,
		Settings: settingsClient,
	}
}

func (rg *RouteGroup) Register(app *fiber.App) {
	app.Post("/spin/superace/deluxe", rg.SpinHandler)
}