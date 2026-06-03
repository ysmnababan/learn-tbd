package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"learn-tbd/config"
	"learn-tbd/internal/app/example_feat"
	"learn-tbd/internal/factory"
)

func Init(e *echo.Echo, f *factory.Factory) {
	cfg := config.Get()

	// index
	e.GET("/", func(c echo.Context) error {
		message := fmt.Sprintf("Welcome to %s", cfg.App.Name)
		return c.String(http.StatusOK, message)
	})

	// routes v1
	api := e.Group("/api/v1")

	example_feat.NewHandler(f).Route(api.Group("/users"))
}
