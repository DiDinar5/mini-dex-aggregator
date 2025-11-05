package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/DiDinar5/mini-dex-aggregator/domain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthereumService struct {
	client              *ethclient.Client
	uniswapV2ABI        abi.ABI
	uniswapV2FactoryABI abi.ABI
	erc20ABI            abi.ABI
	tokenAddresses      map[string]string
	tokenAddressesMu    sync.RWMutex
	tokenInfoCache      map[string]*domain.TokenInfo
	tokenInfoMu         sync.RWMutex
}

const uniswapV2PairABI = `[
	{
		"inputs": [],
		"name": "getReserves",
		"outputs": [
			{"internalType": "uint112", "name": "_reserve0", "type": "uint112"},
			{"internalType": "uint112", "name": "_reserve1", "type": "uint112"},
			{"internalType": "uint32", "name": "_blockTimestampLast", "type": "uint32"}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "token0",
		"outputs": [{"internalType": "address", "name": "", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "token1",
		"outputs": [{"internalType": "address", "name": "", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

const erc20ABI = `[
	{
		"inputs": [],
		"name": "symbol",
		"outputs": [{"internalType": "string", "name": "", "type": "string"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "decimals",
		"outputs": [{"internalType": "uint8", "name": "", "type": "uint8"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

const uniswapV2FactoryABI = `[
	{
		"inputs": [
			{"internalType": "address", "name": "tokenA", "type": "address"},
			{"internalType": "address", "name": "tokenB", "type": "address"}
		],
		"name": "getPair",
		"outputs": [{"internalType": "address", "name": "pair", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

// Known DEX factory addresses on Ethereum mainnet
var dexFactories = map[string]string{
	"UniswapV2": "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f",
	"Sushiswap": "0xC0AEe478e3658e2610c5F7A4A2E1777cE9e4f2Ac",
}

func NewEthereumService(rpcURL string) (*EthereumService, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum: %w", err)
	}

	service := &EthereumService{
		client:         client,
		tokenAddresses: make(map[string]string),
		tokenInfoCache: make(map[string]*domain.TokenInfo),
	}

	if err := service.initABI(); err != nil {
		return nil, fmt.Errorf("failed to initialize ABI: %w", err)
	}

	return service, nil
}

func (e *EthereumService) initABI() error {
	var err error

	e.uniswapV2ABI, err = abi.JSON(strings.NewReader(uniswapV2PairABI))
	if err != nil {
		return fmt.Errorf("failed to parse Uniswap V2 ABI: %w", err)
	}

	e.uniswapV2FactoryABI, err = abi.JSON(strings.NewReader(uniswapV2FactoryABI))
	if err != nil {
		return fmt.Errorf("failed to parse Uniswap V2 Factory ABI: %w", err)
	}

	e.erc20ABI, err = abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	return nil
}

func (e *EthereumService) GetPoolReserves(ctx context.Context, poolAddress string) (*domain.PoolReserves, error) {
	if !common.IsHexAddress(poolAddress) {
		return nil, fmt.Errorf("invalid pool address: %s", poolAddress)
	}

	poolContract := common.HexToAddress(poolAddress)

	e.tokenAddressesMu.RLock()
	cachedAddresses, exists := e.tokenAddresses[poolAddress]
	e.tokenAddressesMu.RUnlock()

	var token0Address, token1Address common.Address
	var err error

	if exists {
		addresses := strings.Split(cachedAddresses, ",")
		if len(addresses) == 2 {
			token0Address = common.HexToAddress(addresses[0])
			token1Address = common.HexToAddress(addresses[1])
		}
	}

	if token0Address == (common.Address{}) || token1Address == (common.Address{}) {
		boundContract := bind.NewBoundContract(poolContract, e.uniswapV2ABI, e.client, e.client, e.client)

		var token0Result []interface{}
		if err := boundContract.Call(&bind.CallOpts{Context: ctx}, &token0Result, "token0"); err != nil {
			return nil, fmt.Errorf("failed to call token0: %w", err)
		}
		if len(token0Result) > 0 {
			if addr, ok := token0Result[0].(common.Address); ok {
				token0Address = addr
			} else {
				return nil, fmt.Errorf("unexpected token0 result type")
			}
		}

		var token1Result []interface{}
		if err := boundContract.Call(&bind.CallOpts{Context: ctx}, &token1Result, "token1"); err != nil {
			return nil, fmt.Errorf("failed to call token1: %w", err)
		}
		if len(token1Result) > 0 {
			if addr, ok := token1Result[0].(common.Address); ok {
				token1Address = addr
			} else {
				return nil, fmt.Errorf("unexpected token1 result type")
			}
		}

		e.tokenAddressesMu.Lock()
		e.tokenAddresses[poolAddress] = token0Address.Hex() + "," + token1Address.Hex()
		e.tokenAddressesMu.Unlock()
	}

	reservesData, err := e.callContract(ctx, poolContract, e.uniswapV2ABI, "getReserves")
	if err != nil {
		return nil, fmt.Errorf("failed to get pool reserves: %w", err)
	}

	var reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}

	if err := e.uniswapV2ABI.UnpackIntoInterface(&reserves, "getReserves", reservesData); err != nil {
		return nil, fmt.Errorf("failed to unpack reserves data: %w", err)
	}

	blockNumber, err := e.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	return &domain.PoolReserves{
		Reserve0:    reserves.Reserve0,
		Reserve1:    reserves.Reserve1,
		Token0:      token0Address.Hex(),
		Token1:      token1Address.Hex(),
		BlockNumber: blockNumber,
	}, nil
}

func (e *EthereumService) GetTokenInfo(ctx context.Context, tokenAddress string) (*domain.TokenInfo, error) {
	if !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("invalid token address: %s", tokenAddress)
	}

	e.tokenInfoMu.RLock()
	if cached, exists := e.tokenInfoCache[tokenAddress]; exists {
		e.tokenInfoMu.RUnlock()
		return cached, nil
	}
	e.tokenInfoMu.RUnlock()

	tokenContract := common.HexToAddress(tokenAddress)
	boundContract := bind.NewBoundContract(tokenContract, e.erc20ABI, e.client, e.client, e.client)

	var symbol string
	var symbolResult []interface{}
	if err := boundContract.Call(&bind.CallOpts{Context: ctx}, &symbolResult, "symbol"); err != nil {
		return nil, fmt.Errorf("failed to call symbol: %w", err)
	}
	if len(symbolResult) > 0 {
		if s, ok := symbolResult[0].(string); ok {
			symbol = s
		} else {
			return nil, fmt.Errorf("unexpected symbol result type")
		}
	}

	var decimals uint8
	var decimalsResult []interface{}
	if err := boundContract.Call(&bind.CallOpts{Context: ctx}, &decimalsResult, "decimals"); err != nil {
		return nil, fmt.Errorf("failed to call decimals: %w", err)
	}

	if len(decimalsResult) > 0 {
		switch v := decimalsResult[0].(type) {
		case uint8:
			decimals = v
		case uint32:
			decimals = uint8(v)
		case uint64:
			decimals = uint8(v)
		case *big.Int:
			if v.IsUint64() && v.Uint64() <= 255 {
				decimals = uint8(v.Uint64())
			} else {
				return nil, fmt.Errorf("decimals value too large: %s", v.String())
			}
		default:
			return nil, fmt.Errorf("unexpected decimals result type: %T", v)
		}
	}

	tokenInfo := &domain.TokenInfo{
		Address:  tokenAddress,
		Symbol:   symbol,
		Decimals: decimals,
	}

	e.tokenInfoMu.Lock()
	e.tokenInfoCache[tokenAddress] = tokenInfo
	e.tokenInfoMu.Unlock()

	return tokenInfo, nil
}

func (e *EthereumService) callContract(ctx context.Context, contract common.Address, parsedABI abi.ABI, method string) ([]byte, error) {
	data, err := parsedABI.Pack(method)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method %s: %w", method, err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract method %s: %w", method, err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty result from contract method %s", method)
	}

	return result, nil
}

// FindPool finds a pool address for a token pair on a specific DEX
func (e *EthereumService) FindPool(ctx context.Context, dexName, tokenA, tokenB string) (string, error) {
	factoryAddress, ok := dexFactories[dexName]
	if !ok {
		return "", fmt.Errorf("unknown DEX: %s", dexName)
	}

	if !common.IsHexAddress(tokenA) || !common.IsHexAddress(tokenB) {
		return "", fmt.Errorf("invalid token address")
	}

	// Ensure token addresses are in correct order (token0 < token1)
	addrA := common.HexToAddress(tokenA)
	addrB := common.HexToAddress(tokenB)

	// Uniswap V2 requires tokens to be in ascending order
	var token0, token1 common.Address
	if addrA.Hex() < addrB.Hex() {
		token0 = addrA
		token1 = addrB
	} else {
		token0 = addrB
		token1 = addrA
	}

	factoryContract := common.HexToAddress(factoryAddress)
	boundContract := bind.NewBoundContract(factoryContract, e.uniswapV2FactoryABI, e.client, e.client, e.client)

	var pairResult []interface{}
	if err := boundContract.Call(&bind.CallOpts{Context: ctx}, &pairResult, "getPair", token0, token1); err != nil {
		return "", fmt.Errorf("failed to call getPair: %w", err)
	}

	if len(pairResult) == 0 {
		return "", fmt.Errorf("pool not found")
	}

	pairAddress, ok := pairResult[0].(common.Address)
	if !ok {
		return "", fmt.Errorf("unexpected pair address type")
	}

	// Check if pool exists (address is not zero)
	if pairAddress == (common.Address{}) {
		return "", fmt.Errorf("pool does not exist")
	}

	return pairAddress.Hex(), nil
}

// GetQuoteForPool calculates the output amount for a given input amount in a pool
func (e *EthereumService) GetQuoteForPool(ctx context.Context, poolAddress, tokenIn string, amountIn *big.Int) (*big.Int, error) {
	poolReserves, err := e.GetPoolReserves(ctx, poolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool reserves: %w", err)
	}

	var reserveIn, reserveOut *big.Int
	if strings.EqualFold(tokenIn, poolReserves.Token0) {
		reserveIn = poolReserves.Reserve0
		reserveOut = poolReserves.Reserve1
	} else if strings.EqualFold(tokenIn, poolReserves.Token1) {
		reserveIn = poolReserves.Reserve1
		reserveOut = poolReserves.Reserve0
	} else {
		return nil, fmt.Errorf("token %s not found in pool", tokenIn)
	}

	// Use Uniswap V2 formula: amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
	// This is the same calculation as in usecase/calculation.go
	feeNumerator := big.NewInt(997)
	feeDenominator := big.NewInt(1000)

	amountInWithFee := new(big.Int).Mul(amountIn, feeNumerator)
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	reserveInWithFee := new(big.Int).Mul(reserveIn, feeDenominator)
	denominator := new(big.Int).Add(reserveInWithFee, amountInWithFee)

	amountOut := new(big.Int).Div(numerator, denominator)

	return amountOut, nil
}

// FindAllPools finds all available pools for a token pair across different DEXes
func (e *EthereumService) FindAllPools(ctx context.Context, tokenA, tokenB string) (map[string]string, error) {
	pools := make(map[string]string)

	for dexName := range dexFactories {
		poolAddress, err := e.FindPool(ctx, dexName, tokenA, tokenB)
		if err != nil {
			// Pool doesn't exist on this DEX, skip
			continue
		}
		pools[dexName] = poolAddress
	}

	return pools, nil
}
