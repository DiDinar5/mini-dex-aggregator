package domain

import "context"

type UsecaseInterface interface {
	Estimate(ctx context.Context, req EstimateRequest) (EstimateResponse, error)
}

type EthereumServiceInterface interface {
	GetPoolReserves(ctx context.Context, poolAddress string) (*PoolReserves, error)
}
