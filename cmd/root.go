package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "trade_game",
	Short: "交易小游戏服务端",
	Long:  "交易小游戏服务端：一个支持买入/卖出/明日价格/特殊商城的多用户交易游戏后端。",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "配置文件路径")
}
