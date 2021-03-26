package tradetools

type Stock struct {
	Ticker   string
	Currency string
}

type TradeClient interface {
	GetStocks() ([]Stock, error)
	AddLimit(limit *Limit) error
	ClearLimits() (int, error)
}
