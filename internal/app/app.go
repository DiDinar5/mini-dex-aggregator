package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DiDinar5/mini-dex-aggregator/config"
	"github.com/DiDinar5/mini-dex-aggregator/infrastructure/ethereum"
	"github.com/DiDinar5/mini-dex-aggregator/internal/handler"
	"github.com/DiDinar5/mini-dex-aggregator/internal/middlewares/validator"
	"github.com/DiDinar5/mini-dex-aggregator/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run(cfg config.Config) {
	ethereumService, err := ethereum.NewEthereumService(cfg.Ethereum.RPCURL)
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum service: %v", err)
	}

	usecaseInstance := usecase.NewUsecase(ethereumService)

	handlerInstance := handler.NewHandler(usecaseInstance)

	e := echo.New()
	e.HideBanner = true

	e.Validator = validator.NewValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	handlerInstance.SetupRoutes(e)

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: e,
	}

	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	gracefulShutdown(server)

}

func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
