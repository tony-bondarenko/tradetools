package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tony-bondarenko/tradetools"
	"github.com/tony-bondarenko/tradetools/quik"
	"github.com/tony-bondarenko/tradetools/tinkoff"
	"log"
)

var configuration = Configuration{}

var rootCmd = &cobra.Command{
	Use:   "trade",
	Short: "Trade is set of commands for trading automation",
	Long:  "Trade is set of commands for trading automation",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&(configuration.cfgFile), "config", "c", "", "config file (default is $HOME/.trade.json)")
	rootCmd.PersistentFlags().StringVarP(&(configuration.providerName), "provider", "p", "", "provider to use")
	_ = rootCmd.MarkPersistentFlagRequired("provider")
	cobra.OnInitialize(func() {
		err := configuration.initConfig()
		if err != nil {
			log.Fatalln(err)
		}
	})
}

func createTradeClient() (tradetools.TradeClient, error) {
	switch configuration.providerName {
	case "tinkoff":
		return tinkoff.CreateClient(configuration.providerConfig)

	case "open":
		return quik.CreateClient(configuration.providerConfig)
	}
	return nil, fmt.Errorf("unknown provider: %s", configuration.providerName)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
