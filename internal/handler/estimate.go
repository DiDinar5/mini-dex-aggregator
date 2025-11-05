package handler

import (
	"net/http"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
	"github.com/labstack/echo/v4"
)

func (h *Handler) EstimateHandler(c echo.Context) error {
	var req domain.EstimateRequest

	if err := echo.QueryParamsBinder(c).
		String("pool", &req.Pool).
		String("src", &req.Src).
		String("dst", &req.Dst).
		String("src_amount", &req.SrcAmount).
		BindError(); err != nil {
		errrorJson(http.StatusBadRequest, err.Error(), c.Response().Writer)
		return nil
	}

	if err := c.Validate(&req); err != nil {
		errrorJson(http.StatusBadRequest, err.Error(), c.Response().Writer)
		return nil
	}

	response, err := h.usecase.Estimate(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:       "Estimation failed",
			Code:        http.StatusInternalServerError,
			Description: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
