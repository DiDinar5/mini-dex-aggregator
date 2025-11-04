package handler

import (
	"github.com/DiDinar5/1inch_test_task/domain"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	usecase domain.UsecaseInterface
}

func NewHandler(usecase domain.UsecaseInterface) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/estimate", h.EstimateHandler)
}
