package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(tickerCmd)
}

var tickerCmd = &cobra.Command{
	Use:   "tickers",
	Short: "Print the list of tickers",
	Long:  `Print the list of tickers`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		tradeClient, err := createTradeClient()
		if err != nil {
			log.Fatalln(err)
		}

		stocks, err := tradeClient.GetStocks()
		if err != nil {
			log.Fatalln(err)
		}

		for _, stock := range stocks {
			fmt.Println(stock.Ticker)
		}
	},
}
