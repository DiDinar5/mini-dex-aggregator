package usecase

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
	"github.com/DiDinar5/mini-dex-aggregator/infrastructure/ethereum"
)

type QuoteUsecase struct {
	ethereumService domain.EthereumServiceInterface
	graphService    domain.TheGraphServiceInterface
	minTVL          float64
}

func NewQuoteUsecase(ethereumService domain.EthereumServiceInterface, graphService domain.TheGraphServiceInterface, minTVL float64) *QuoteUsecase {
	return &QuoteUsecase{
		ethereumService: ethereumService,
		graphService:    graphService,
		minTVL:          minTVL,
	}
}

func (u *QuoteUsecase) Quote(ctx context.Context, req domain.QuoteRequest) (domain.QuoteResponse, error) {
	fromSymbol := strings.ToUpper(req.From)
	toSymbol := strings.ToUpper(req.To)

	if fromSymbol == "ETH" {
		fromSymbol = "WETH"
	}
	if toSymbol == "ETH" {
		toSymbol = "WETH"
	}

	fromTokenAddr := ethereum.GetTokenAddress(fromSymbol)
	if fromTokenAddr == "" {
		return domain.QuoteResponse{}, fmt.Errorf("unknown token symbol: %s", req.From)
	}

	toTokenAddr := ethereum.GetTokenAddress(toSymbol)
	if toTokenAddr == "" {
		return domain.QuoteResponse{}, fmt.Errorf("unknown token symbol: %s", req.To)
	}

	amountIn, err := u.parseAmount(req.Amount)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to parse amount: %w", err)
	}

	fromTokenInfo, err := u.ethereumService.GetTokenInfo(ctx, fromTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get from token info: %w", err)
	}

	toTokenInfo, err := u.ethereumService.GetTokenInfo(ctx, toTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get to token info: %w", err)
	}

	amountInWei := u.adjustForDecimals(amountIn, fromTokenInfo.Decimals)

	pools, err := u.ethereumService.FindAllPools(ctx, fromTokenAddr, toTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to find pools: %w", err)
	}

	if len(pools) == 0 {
		return domain.QuoteResponse{}, fmt.Errorf("no pools found for pair %s/%s", req.From, req.To)
	}

	poolDataMap := make(map[string]*domain.PoolData)
	if u.graphService != nil {
		graphPools, err := u.graphService.GetPoolsByTokenPair(ctx, fromTokenAddr, toTokenAddr)
		if err == nil {
			for _, poolData := range graphPools {
				poolDataMap[strings.ToLower(poolData.ID)] = poolData
			}
		}
		for _, poolAddress := range pools {
			poolLower := strings.ToLower(poolAddress)
			if _, exists := poolDataMap[poolLower]; !exists {
				if poolData, err := u.graphService.GetPoolData(ctx, poolAddress); err == nil {
					poolDataMap[poolLower] = poolData
				}
			}
		}
	}

	var allQuotes []domain.DEXQuote
	var bestQuote *domain.DEXQuote
	var bestAmount *big.Int

	for dexName, poolAddress := range pools {
		poolLower := strings.ToLower(poolAddress)
		poolData := poolDataMap[poolLower]

		if poolData != nil && poolData.ReserveUSD < u.minTVL {
			continue
		}

		amountOut, err := u.ethereumService.GetQuoteForPool(ctx, poolAddress, fromTokenAddr, amountInWei)
		if err != nil {
			continue
		}

		amountOutAdjusted := u.adjustFromDecimals(amountOut, toTokenInfo.Decimals)

		quote := domain.DEXQuote{
			DEX:      dexName,
			Pool:     poolAddress,
			ToAmount: amountOutAdjusted.String(),
		}

		if poolData != nil {
			quote.PoolInfo = u.buildPoolInfo(poolData)
		}

		allQuotes = append(allQuotes, quote)

		if bestAmount == nil || amountOut.Cmp(bestAmount) > 0 {
			bestAmount = amountOut
			bestQuote = &quote
		}
	}

	if bestQuote == nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get quotes from any pool")
	}

	bestAmountAdjusted := u.adjustFromDecimals(bestAmount, toTokenInfo.Decimals)

	response := domain.QuoteResponse{
		FromToken:  req.From,
		ToToken:    req.To,
		FromAmount: req.Amount,
		ToAmount:   bestAmountAdjusted.String(),
		BestQuote:  *bestQuote,
		AllQuotes:  allQuotes,
	}

	return response, nil
}

func (u *QuoteUsecase) parseAmount(amountStr string) (*big.Int, error) {
	amountStr = strings.TrimSpace(amountStr)

	if amountStr == "" {
		return nil, fmt.Errorf("amount cannot be empty")
	}

	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	return amount, nil
}

func (u *QuoteUsecase) adjustForDecimals(amount *big.Int, decimals uint8) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Mul(amount, multiplier)
}

func (u *QuoteUsecase) adjustFromDecimals(amount *big.Int, decimals uint8) *big.Int {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Div(amount, divisor)
}

func (u *QuoteUsecase) buildPoolInfo(poolData *domain.PoolData) *domain.PoolInfo {
	isActive := poolData.ReserveUSD >= u.minTVL

	return &domain.PoolInfo{
		TVL:          fmt.Sprintf("%.2f", poolData.ReserveUSD),
		Volume24h:    fmt.Sprintf("%.2f", poolData.Volume24hUSD),
		Fees24h:      fmt.Sprintf("%.2f", poolData.Fees24hUSD),
		Reserve0:     poolData.Reserve0,
		Reserve1:     poolData.Reserve1,
		Token0Symbol: poolData.Token0Symbol,
		Token1Symbol: poolData.Token1Symbol,
		IsActive:     isActive,
	}
}

func (u *QuoteUsecase) formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
