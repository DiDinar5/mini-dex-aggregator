package ethereum

var tokenAddresses = map[string]string{
	"ETH":  "0x0000000000000000000000000000000000000000",
	"WETH": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
	"USDC": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
	"USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	"DAI":  "0x6B175474E89094C44Da98b954EedeAC495271d0F",
	"WBTC": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
	"UNI":  "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
}

// GetTokenAddress returns the token address for a given symbol
// Returns empty string if token is not found
func GetTokenAddress(symbol string) string {
	if addr, ok := tokenAddresses[symbol]; ok {
		return addr
	}
	return ""
}

// IsValidTokenSymbol checks if a token symbol is known
func IsValidTokenSymbol(symbol string) bool {
	_, ok := tokenAddresses[symbol]
	return ok
}
