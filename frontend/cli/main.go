// trade_game_cli 交易小游戏命令行客户端
package main

import (
	"log"

	"github.com/spf13/cobra"
)

var serverAddr string

var rootCmd = &cobra.Command{
	Use:          "trade_game_cli",
	Short:        "交易小游戏命令行客户端",
	Long:         "交易小游戏命令行客户端：连接 trade_game 服务端进行登录/注册、交易、推进明天、购买仓库空间等操作。",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewApp(serverAddr).Run()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&serverAddr, "server", "s", "http://127.0.0.1:8080", "服务端地址（如 http://127.0.0.1:8080）")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
