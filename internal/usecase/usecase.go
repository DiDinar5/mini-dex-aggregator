package usecase

import (
	"github.com/DiDinar5/1inch_test_task/domain"
)

func NewUsecase() domain.UsecaseInterface {
	return NewEstimateUsecase()
}
