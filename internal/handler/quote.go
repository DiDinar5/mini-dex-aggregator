package handler

import (
	"net/http"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
	"github.com/labstack/echo/v4"
)

func (h *Handler) QuoteHandler(c echo.Context) error {
	var req domain.QuoteRequest

	if err := echo.QueryParamsBinder(c).
		String("from", &req.From).
		String("to", &req.To).
		String("amount", &req.Amount).
		BindError(); err != nil {
		errrorJson(http.StatusBadRequest, err.Error(), c.Response().Writer)
		return nil
	}

	if err := c.Validate(&req); err != nil {
		errrorJson(http.StatusBadRequest, err.Error(), c.Response().Writer)
		return nil
	}

	response, err := h.usecase.Quote(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:       "Quote failed",
			Code:        http.StatusInternalServerError,
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
