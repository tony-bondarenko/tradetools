package tinkoff

import (
	"context"
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/tony-bondarenko/tradetools"
	"strings"
	"time"
)

const (
	requestTimeout       = 5 * time.Second
	rateLimitHttpCode    = 429
	rateLimitWaitSeconds = 30
	rateLimitWaitTime    = rateLimitWaitSeconds * time.Second
)

type TradeClient struct {
	client  *sdk.RestClient
	config  *ClientConfiguration
	figiMap map[string]string
}

func CreateClient(configuration interface{}) (*TradeClient, error) {
	config, err := createClientConfig(configuration)
	if err != nil {
		return nil, err
	}
	tradeClient := new(TradeClient)
	tradeClient.client = sdk.NewRestClient(config.token)
	tradeClient.config = config
	return tradeClient, nil
}

func (t *TradeClient) GetStocks() ([]tradetools.Stock, error) {
	instruments, err := t.getInstruments()
	if err != nil {
		return nil, err
	}

	var stocks = make([]tradetools.Stock, 0)
	for _, instrument := range instruments {
		if instrument.Currency == "USD" {
			stocks = append(stocks, tradetools.Stock{
				Ticker:   instrument.Ticker,
				Currency: string(instrument.Currency),
			})
		}
	}
	return stocks, nil
}

func (t *TradeClient) getInstruments() ([]sdk.Instrument, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	instruments, err := t.client.Stocks(ctx)
	if err != nil {
		return nil, err
	}

	return instruments, err
}

func (t *TradeClient) AddLimit(limit *tradetools.Limit) error {
	figi, err := t.getFigiByTicker(limit.Ticker)
	if err != nil {
		return err
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		_, err = t.client.LimitOrder(ctx, sdk.DefaultAccount, figi, limit.Lots, sdk.BUY, limit.Price)
		cancel()
		if t.isRateLimitError(err) {
			t.waitRateLimit()
		} else {
			break
		}
	}
	return err
}

func (t *TradeClient) ClearLimits() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	orders, err := t.client.Orders(ctx, sdk.DefaultAccount)
	if err != nil {
		return 0, err
	}

	orderNum := 0
	for _, order := range orders {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = t.client.OrderCancel(ctx, sdk.DefaultAccount, order.ID)
			cancel()

			if t.isRateLimitError(err) {
				fmt.Printf("%d limits has been cleared", orderNum)
				t.waitRateLimit()
			} else {
				break
			}
		}

		if err != nil {
			return 0, err
		}
		orderNum++
	}
	return orderNum, nil
}

func (t *TradeClient) isRateLimitError(err error) bool {
	return err != nil && strings.Index(err.Error(), fmt.Sprintf("code=%d", rateLimitHttpCode)) != -1
}

func (t *TradeClient) waitRateLimit() {
	fmt.Println(fmt.Sprintf("Waiting for %d seconds...", rateLimitWaitSeconds))
	time.Sleep(rateLimitWaitTime)
	fmt.Println("Continue...")
}

func (t *TradeClient) getFigiByTicker(ticker string) (string, error) {
	if t.figiMap == nil {
		t.figiMap = make(map[string]string)
		instruments, err := t.getInstruments()
		if err != nil {
			return "", err
		}
		for _, stock := range instruments {
			t.figiMap[stock.Ticker] = stock.FIGI
		}
	}

	figi, ok := t.figiMap[ticker]
	if !ok {
		return "", fmt.Errorf("unknown ticker: %s", figi)
	}

	return figi, nil
}
