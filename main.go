package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"learn-tbd/config"
	"learn-tbd/internal/factory"
	middleware "learn-tbd/internal/middleware"
	"learn-tbd/internal/pkg/database"
	"learn-tbd/internal/pkg/logger"
	httpserver "learn-tbd/internal/server"
	"learn-tbd/internal/utils/env"
)

func init() {
	selectedEnv := config.Env()
	env := env.NewEnv()
	env.Load(`.env`)
	logger.InitLogger()
	log.Info().Msg("Choosen environment " + selectedEnv)
}

func main() {
	cfg := config.Get()

	port := cfg.App.Port

	logLevel, err := zerolog.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	database.Init("std")

	f := factory.NewFactory()

	e := echo.New()
	e.HideBanner = true
	e.IPExtractor = echo.ExtractIPDirect()
	middleware.Init(e)
	httpserver.Init(e, f)

	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}
}
