package main

import (
	"github.com/spf13/cobra"
	"github.com/tony-bondarenko/tradetools/xlsx"
	"log"
)

func init() {
	rootCmd.AddCommand(limitCmd)
	limitCmd.AddCommand(limitSetCmd)
	limitCmd.AddCommand(limitClearCmd)
}

var limitCmd = &cobra.Command{
	Use:   "limit",
	Short: "Set limits based on xlsx template",
	Long:  `Set limits based on xlsx template`,
	Args:  cobra.ExactArgs(1),
	Run:   func(cmd *cobra.Command, args []string) {},
}

var limitSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set limits based on xlsx template",
	Long:  `Set limits based on xlsx template`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tradeClient, err := createTradeClient()
		if err != nil {
			log.Fatalln(err)
		}

		reader, err := xlsx.CreateLimitReader(args[0])
		if err != nil {
			log.Fatalln(err)
		}

		limitNum := 0
		for {
			limit, err := reader.NextLimit()
			if err != nil {
				log.Fatalln(err)
			}

			if limit == nil {
				break
			}

			err = tradeClient.AddLimit(limit)
			if err != nil {
				log.Fatalln(err)
			}
			limitNum++

			if limitNum%10 == 0 {
				log.Printf("%d limits are set so far\n", limitNum)
			}
		}
		log.Printf("%d limits are set\n", limitNum)
	},
}

var limitClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear limits",
	Long:  "Clear limits",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		tradeClient, err := createTradeClient()
		if err != nil {
			log.Fatalln(err)
		}

		limitNum, err := tradeClient.ClearLimits()
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("%d limits has been cleared\n", limitNum)
	},
}
