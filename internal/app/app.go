package app

import (
	"github.com/DiDinar5/mini-dex-aggregator/config"
	"github.com/DiDinar5/mini-dex-aggregator/internal/handler"
	"github.com/DiDinar5/mini-dex-aggregator/internal/usecase"
	"github.com/labstack/echo/v4"
)

func Run(cfg config.Config) {
	e := echo.New()
	usecaseInstance := usecase.NewUsecase()

	handlerInstance := handler.NewHandler(usecaseInstance)

	handlerInstance.SetupRoutes(e)

	e.Start(cfg.Server.Port)
}
