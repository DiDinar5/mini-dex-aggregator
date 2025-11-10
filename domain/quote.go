package domain

type QuoteRequest struct {
	From   string `json:"from" validate:"required"`
	To     string `json:"to" validate:"required"`
	Amount string `json:"amount" validate:"required"`
}

type QuoteResponse struct {
	FromToken  string     `json:"from_token"`
	ToToken    string     `json:"to_token"`
	FromAmount string     `json:"from_amount"`
	ToAmount   string     `json:"to_amount"`
	BestQuote  DEXQuote   `json:"best_quote"`
	AllQuotes  []DEXQuote `json:"all_quotes"`
}

type DEXQuote struct {
	DEX      string    `json:"dex"`
	Pool     string    `json:"pool"`
	ToAmount string    `json:"to_amount"`
	Price    string    `json:"price,omitempty"`
	PoolInfo *PoolInfo `json:"pool_info,omitempty"`
}

type PoolInfo struct {
	TVL          string `json:"tvl"`
	Volume24h    string `json:"volume_24h"`
	Fees24h      string `json:"fees_24h"`
	Reserve0     string `json:"reserve0"`
	Reserve1     string `json:"reserve1"`
	Token0Symbol string `json:"token0_symbol"`
	Token1Symbol string `json:"token1_symbol"`
	IsActive     bool   `json:"is_active"`
}
