package tradetools

type Limit struct {
	Ticker string
	Price  float64
	Lots   int
}

type LimitReader interface {
	NextLimit() (*Limit, error)
}
