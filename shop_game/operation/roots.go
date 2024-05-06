package operation

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

/*
cobra相关内容
*/

var configFilePath string
var rootCmd = cobra.Command{
	Use:   "config",
	Short: "input config file address.",
	Run: func(cmd *cobra.Command, args []string) {
		if checkConfigFile(configFilePath) {
			// 启动程序 start
			NewStart(configFilePath)
		}

	},
}
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v1.0")
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "used for initializing the environment for the first time.",
	Run: func(cmd *cobra.Command, args []string) {
		if checkConfigFile(configFilePath) {
			initEnvironment(configFilePath)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path.")
	rootCmd.AddCommand(versionCmd)
	//NewDB(&DB)
	initCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path.")
	rootCmd.AddCommand(initCmd)
}

func checkConfigFile(configFilePath string) bool {
	if configFilePath == "" {
		fmt.Println("please input --config!")
		return false
	}
	fmt.Println("start!Use config file is :", configFilePath)
	return true
}

func Start() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln("start error! please check database config!")
	}
}

// go get -u github.com/spf13/cobra
