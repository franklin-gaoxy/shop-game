package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"trade_game/internal/config"
	"trade_game/internal/database"
	_ "trade_game/internal/database/mysql" // 注册 MySQL 驱动
)

var (
	forceInit   bool
	sqlFile     string
	dataFile    string
	initialKey  string
	initialUses int
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化数据库（建表、导入商品/暴击/默认配置、创建管理员与初始密钥）",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("读取配置失败: %w", err)
		}

		// 读取初始化数据 yaml（商品、暴击事件、默认仓库大小等）
		raw, err := os.ReadFile(dataFile)
		if err != nil {
			return fmt.Errorf("读取初始化数据文件失败: %w", err)
		}
		var data database.InitData
		if err := yaml.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("解析初始化数据文件失败: %w", err)
		}
		// 密钥参数（可选，未指定则随机生成）
		if initialUses > 0 {
			data.Defaults.InitialKeyMaxUses = initialUses
		}

		db, err := database.New(cfg.Config.Database)
		if err != nil {
			return err
		}
		defer db.Close()

		// 建表：若已初始化且未指定 --force 则跳过
		if _, err := db.Initialize(sqlFile, forceInit); err != nil {
			if errors.Is(err, database.ErrAlreadyInitialized) {
				log.Println("数据库已初始化，跳过初始化。如需重新初始化请加 --force 参数（会删除已有数据）")
				return nil
			}
			return err
		}

		// 生成初始密钥（建表之后，保证唯一性校验可用）
		if initialKey == "" {
			initialKey, err = db.GenerateKey()
			if err != nil {
				return err
			}
		}
		data.Defaults.InitialKey = initialKey

		// 导入数据：管理员、密钥、商品、暴击事件、默认配置
		key, err := db.ImportInitData(&data, cfg.Config.Platform.Username, cfg.Config.Platform.Password)
		if err != nil {
			return err
		}

		log.Println("========== 初始化完成 ==========")
		log.Printf("管理员用户名: %s", cfg.Config.Platform.Username)
		log.Printf("初始注册密钥: %s", key)
		log.Printf("商品数量: %d，暴击事件数量: %d", len(data.Products), len(data.CritEvents))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&forceInit, "force", false, "强制重新初始化（删除已有数据后重建）")
	initCmd.Flags().StringVar(&sqlFile, "sql", "sql/init.sql", "建表 SQL 文件路径")
	initCmd.Flags().StringVar(&dataFile, "data", "data/init_data.yaml", "初始化数据 yaml 文件路径")
	initCmd.Flags().StringVar(&initialKey, "key", "", "指定初始注册密钥（默认随机生成 10 位）")
	initCmd.Flags().IntVar(&initialUses, "key-uses", 0, "初始密钥可用次数（默认取数据文件配置或 10）")
}
