package thegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
)

type TheGraphService struct {
	client       *http.Client
	uniswapV2URL string
	minTVL       float64
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type PairResponse struct {
	Pair *PairData `json:"pair"`
}

type PairsResponse struct {
	Pairs []*PairData `json:"pairs"`
}

type PairData struct {
	ID           string `json:"id"`
	Token0       Token  `json:"token0"`
	Token1       Token  `json:"token1"`
	Reserve0     string `json:"reserve0"`
	Reserve1     string `json:"reserve1"`
	TotalSupply  string `json:"totalSupply"`
	ReserveUSD   string `json:"reserveUSD"`
	VolumeUSD    string `json:"volumeUSD"`
	Volume24hUSD string `json:"volumeUSD24h,omitempty"`
	Fees24hUSD   string `json:"feesUSD24h,omitempty"`
}

type Token struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
}

func NewTheGraphService(uniswapV2URL string, minTVL float64) *TheGraphService {
	return &TheGraphService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		uniswapV2URL: uniswapV2URL,
		minTVL:       minTVL,
	}
}

func (s *TheGraphService) GetPoolData(ctx context.Context, poolAddress string) (*domain.PoolData, error) {
	query := `
	query GetPair($id: ID!) {
		pair(id: $id) {
			id
			token0 { id symbol }
			token1 { id symbol }
			reserve0
			reserve1
			totalSupply
			reserveUSD
			volumeUSD
		}
	}`

	vars := map[string]interface{}{
		"id": strings.ToLower(poolAddress),
	}

	var resp PairResponse
	if err := s.executeQuery(ctx, query, vars, &resp); err != nil {
		return nil, fmt.Errorf("GetPoolData failed: %w", err)
	}

	if resp.Pair == nil {
		return nil, fmt.Errorf("pool not found: %s", poolAddress)
	}

	return s.convertPairToPoolData(resp.Pair), nil
}

func (s *TheGraphService) GetPoolsByTokenPair(ctx context.Context, token0, token1 string) ([]*domain.PoolData, error) {
	token0Lower := strings.ToLower(token0)
	token1Lower := strings.ToLower(token1)

	query := `
	query GetPairs($token0: String!, $token1: String!) {
		pairs(
			where: {
				_or: [
					{ token0: $token0, token1: $token1 },
					{ token0: $token1, token1: $token0 }
				]
			},
			orderBy: reserveUSD,
			orderDirection: desc,
			first: 10
		) {
			id
			token0 {
				id
				symbol
			}
			token1 {
				id
				symbol
			}
			reserve0
			reserve1
			totalSupply
			reserveUSD
			volumeUSD
		}
	}
`

	vars := map[string]interface{}{
		"token0": token0Lower,
		"token1": token1Lower,
	}

	var resp PairsResponse
	if err := s.executeQuery(ctx, query, vars, &resp); err != nil {
		return nil, fmt.Errorf("GetPoolsByTokenPair failed: %w", err)
	}

	var pools []*domain.PoolData
	for _, p := range resp.Pairs {
		poolData := s.convertPairToPoolData(p)
		if poolData.ReserveUSD >= s.minTVL {
			pools = append(pools, poolData)
		}
	}

	return pools, nil
}

func (s *TheGraphService) executeQuery(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.uniswapV2URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors)
	}

	if err := json.Unmarshal(graphQLResp.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal GraphQL data: %w", err)
	}

	return nil
}

func (s *TheGraphService) convertPairToPoolData(pair *PairData) *domain.PoolData {
	reserveUSD, _ := strconv.ParseFloat(pair.ReserveUSD, 64)
	volumeUSD, _ := strconv.ParseFloat(pair.VolumeUSD, 64)

	volume24hUSD := 0.0
	if pair.Volume24hUSD != "" {
		volume24hUSD, _ = strconv.ParseFloat(pair.Volume24hUSD, 64)
	}

	fees24hUSD := 0.0
	if pair.Fees24hUSD != "" {
		fees24hUSD, _ = strconv.ParseFloat(pair.Fees24hUSD, 64)
	}

	return &domain.PoolData{
		ID:           pair.ID,
		Token0:       pair.Token0.ID,
		Token1:       pair.Token1.ID,
		Reserve0:     pair.Reserve0,
		Reserve1:     pair.Reserve1,
		TotalSupply:  pair.TotalSupply,
		ReserveUSD:   reserveUSD,
		VolumeUSD:    volumeUSD,
		Volume24hUSD: volume24hUSD,
		Fees24hUSD:   fees24hUSD,
		Token0Symbol: pair.Token0.Symbol,
		Token1Symbol: pair.Token1.Symbol,
	}
}
