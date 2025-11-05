package usecase

import "github.com/DiDinar5/mini-dex-aggregator/domain"

func NewUsecase(ethereumService domain.EthereumServiceInterface) domain.UsecaseInterface {
	return NewEstimateUsecase(ethereumService)
}

const (
	UniswapV2FeeNumerator   = 997
	UniswapV2FeeDenominator = 1000
)
