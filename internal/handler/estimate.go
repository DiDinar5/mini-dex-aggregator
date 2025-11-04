package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) EstimateHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "test")
}
