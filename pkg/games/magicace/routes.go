package magicace

import (
	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/settings"
	"github.com/gofiber/fiber/v2"
)

// RouteGroup holds the dependencies for the handlers
type RouteGroup struct {
    RNG     *rng.Client
    Settings *settings.Client
}

// NewRouteGroup creates a new RouteGroup
func NewRouteGroup(rngClient *rng.Client, settingsClient *settings.Client) *RouteGroup {
    return &RouteGroup{
        RNG:     rngClient,
        Settings: settingsClient,
    }
}

// Register registers the routes with the Fiber app
func (rg *RouteGroup) Register(app *fiber.App) {
    app.Post("/spin/magicace", rg.SpinHandler)
    app.Post("/cascade/magicace", rg.CascadeHandler)
    app.Post("/featureBuy/magicace", rg.FeatureBuyHandler)
}