package domain

import "context"

type TheGraphServiceInterface interface {
	GetPoolData(ctx context.Context, poolAddress string) (*PoolData, error)
	GetPoolsByTokenPair(ctx context.Context, token0, token1 string) ([]*PoolData, error)
}

type PoolData struct {
	ID           string  `json:"id"`
	Token0       string  `json:"token0"`
	Token1       string  `json:"token1"`
	Reserve0     string  `json:"reserve0"`
	Reserve1     string  `json:"reserve1"`
	TotalSupply  string  `json:"totalSupply"`
	ReserveUSD   float64 `json:"reserveUSD"`
	VolumeUSD    float64 `json:"volumeUSD"`
	Volume24hUSD float64 `json:"volume24hUSD"`
	Fees24hUSD   float64 `json:"fees24hUSD"`
	Token0Symbol string  `json:"token0Symbol"`
	Token1Symbol string  `json:"token1Symbol"`
}
