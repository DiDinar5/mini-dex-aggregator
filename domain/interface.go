package domain

import (
	"context"
	"math/big"
)

type UsecaseInterface interface {
	Estimate(ctx context.Context, req EstimateRequest) (EstimateResponse, error)
	Quote(ctx context.Context, req QuoteRequest) (QuoteResponse, error)
}

type EthereumServiceInterface interface {
	GetPoolReserves(ctx context.Context, poolAddress string) (*PoolReserves, error)
	GetTokenInfo(ctx context.Context, tokenAddress string) (*TokenInfo, error)
	FindPool(ctx context.Context, dexName, tokenA, tokenB string) (string, error)
	GetQuoteForPool(ctx context.Context, poolAddress, tokenIn string, amountIn *big.Int) (*big.Int, error)
	FindAllPools(ctx context.Context, tokenA, tokenB string) (map[string]string, error)
}
