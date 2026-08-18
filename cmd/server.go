package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"trade_game/internal/config"
	"trade_game/internal/database"
	_ "trade_game/internal/database/mysql" // 注册 MySQL 驱动
	"trade_game/internal/handler"
)

var serverPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 HTTP 服务",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}

		// 连接数据库并检查运行环境
		db, err := database.New(cfg.Config.Database)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			return fmt.Errorf("数据库连接检查失败: %w", err)
		}
		log.Println("数据库连接成功")

		initialized, err := db.IsInitialized()
		if err != nil {
			return fmt.Errorf("检查初始化状态失败: %w", err)
		}
		if !initialized {
			return fmt.Errorf("数据库尚未初始化，请先执行 init 子命令（trade_game init）")
		}
		log.Println("环境检查通过，数据库已初始化")

		port := serverPort
		if port == 0 {
			port = cfg.Config.Server.Port
		}
		addr := fmt.Sprintf("%s:%d", cfg.Config.Server.Host, port)
		router := handler.NewRouter(db)
		log.Printf("HTTP 服务启动: http://%s\n", addr)
		return router.Run(addr)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&serverPort, "port", 0, "服务监听端口（默认读取配置文件）")
}
