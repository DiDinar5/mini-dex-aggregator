package usecase

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
	"github.com/DiDinar5/mini-dex-aggregator/infrastructure/ethereum"
)

type QuoteUsecase struct {
	ethereumService domain.EthereumServiceInterface
}

func NewQuoteUsecase(ethereumService domain.EthereumServiceInterface) *QuoteUsecase {
	return &QuoteUsecase{
		ethereumService: ethereumService,
	}
}

func (u *QuoteUsecase) Quote(ctx context.Context, req domain.QuoteRequest) (domain.QuoteResponse, error) {
	// Get token addresses from symbols
	fromSymbol := strings.ToUpper(req.From)
	toSymbol := strings.ToUpper(req.To)

	// Handle ETH -> WETH conversion early
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

	// Parse amount
	amountIn, err := u.parseAmount(req.Amount)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Get token info for decimals
	fromTokenInfo, err := u.ethereumService.GetTokenInfo(ctx, fromTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get from token info: %w", err)
	}

	toTokenInfo, err := u.ethereumService.GetTokenInfo(ctx, toTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get to token info: %w", err)
	}

	// Adjust amount for token decimals
	amountInWei := u.adjustForDecimals(amountIn, fromTokenInfo.Decimals)

	// Find all available pools
	pools, err := u.ethereumService.FindAllPools(ctx, fromTokenAddr, toTokenAddr)
	if err != nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to find pools: %w", err)
	}

	if len(pools) == 0 {
		return domain.QuoteResponse{}, fmt.Errorf("no pools found for pair %s/%s", req.From, req.To)
	}

	// Get quotes from all pools
	var allQuotes []domain.DEXQuote
	var bestQuote *domain.DEXQuote
	var bestAmount *big.Int

	for dexName, poolAddress := range pools {
		amountOut, err := u.ethereumService.GetQuoteForPool(ctx, poolAddress, fromTokenAddr, amountInWei)
		if err != nil {
			// Skip pools that fail
			continue
		}

		// Adjust amount out for token decimals
		amountOutAdjusted := u.adjustFromDecimals(amountOut, toTokenInfo.Decimals)

		quote := domain.DEXQuote{
			DEX:      dexName,
			Pool:     poolAddress,
			ToAmount: amountOutAdjusted.String(),
		}

		allQuotes = append(allQuotes, quote)

		// Track best quote (highest output amount)
		if bestAmount == nil || amountOut.Cmp(bestAmount) > 0 {
			bestAmount = amountOut
			bestQuote = &quote
		}
	}

	if bestQuote == nil {
		return domain.QuoteResponse{}, fmt.Errorf("failed to get quotes from any pool")
	}

	// Adjust best quote amount for display
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

// adjustForDecimals converts human-readable amount to wei (multiplies by 10^decimals)
func (u *QuoteUsecase) adjustForDecimals(amount *big.Int, decimals uint8) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Mul(amount, multiplier)
}

// adjustFromDecimals converts wei to human-readable amount (divides by 10^decimals)
func (u *QuoteUsecase) adjustFromDecimals(amount *big.Int, decimals uint8) *big.Int {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Div(amount, divisor)
}
