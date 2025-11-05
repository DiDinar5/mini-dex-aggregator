package usecase

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
)

type EstimateUsecase struct {
	ethereumService domain.EthereumServiceInterface
}

func NewEstimateUsecase(ethereumService domain.EthereumServiceInterface) *EstimateUsecase {
	return &EstimateUsecase{
		ethereumService: ethereumService,
	}
}

func (u *EstimateUsecase) Estimate(ctx context.Context, req domain.EstimateRequest) (domain.EstimateResponse, error) {
	poolReserves, err := u.ethereumService.GetPoolReserves(ctx, req.Pool)
	if err != nil {
		return domain.EstimateResponse{}, fmt.Errorf("failed to get pool reserves: %w", err)
	}

	srcAmount, err := u.parseAmount(req.SrcAmount)
	if err != nil {
		return domain.EstimateResponse{}, fmt.Errorf("failed to parse source amount: %w", err)
	}

	var reserveIn, reserveOut *big.Int
	if strings.EqualFold(req.Src, poolReserves.Token0) {
		reserveIn = poolReserves.Reserve0
		reserveOut = poolReserves.Reserve1
	} else {
		reserveIn = poolReserves.Reserve1
		reserveOut = poolReserves.Reserve0
	}

	dstAmount, err := u.calculateAMMOutput(srcAmount, reserveIn, reserveOut)
	if err != nil {
		return domain.EstimateResponse{}, fmt.Errorf("failed to calculate AMM output: %w", err)
	}

	dstAmountStr := dstAmount.String()

	return domain.EstimateResponse{
		DstAmount: dstAmountStr,
	}, nil
}
