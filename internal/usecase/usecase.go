package usecase

import (
	"context"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
)

type CombinedUsecase struct {
	estimateUsecase *EstimateUsecase
	quoteUsecase    *QuoteUsecase
}

func (c *CombinedUsecase) Estimate(ctx context.Context, req domain.EstimateRequest) (domain.EstimateResponse, error) {
	return c.estimateUsecase.Estimate(ctx, req)
}

func (c *CombinedUsecase) Quote(ctx context.Context, req domain.QuoteRequest) (domain.QuoteResponse, error) {
	return c.quoteUsecase.Quote(ctx, req)
}

func NewUsecase(ethereumService domain.EthereumServiceInterface, graphService domain.TheGraphServiceInterface, minTVL float64) domain.UsecaseInterface {
	return &CombinedUsecase{
		estimateUsecase: NewEstimateUsecase(ethereumService),
		quoteUsecase:    NewQuoteUsecase(ethereumService, graphService, minTVL),
	}
}

const (
	UniswapV2FeeNumerator   = 997
	UniswapV2FeeDenominator = 1000
)
