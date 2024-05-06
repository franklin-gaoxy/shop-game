package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type ConfigFile struct {
	Config struct {
		Database struct { // database information
			Type     string `yaml:"type"`
			Database string `yaml:"database"`
			Address  string `yaml:"address"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"database"`
		Goods struct {
			Import_file string `yaml:"import_file"` // product information file.
		} `yaml:"goods"`
		Difficulty_level struct {
			Probability_of_price_increase int `yaml:"probability_of_price_increase"` // game mode.simple or diffcult
		} `yaml:"difficulty_level"`
		Login_method             loginConfig `yaml:"login_method"`
		Initialize_configuration struct {
			Default_money         int64           `yaml:"default_money"` // default amount of money
			Warehouse             warehouseConfig `yaml:"warehouse"`
			Refrigerate_warehouse warehouseConfig `yaml:"refrigerate_warehouse"`
		} `yaml:"initialize_configuration"`
	} `yaml:"config"`
}

type warehouseConfig struct {
	// 此为 默认仓库大小 和 扩容1点仓库所需金额 不需要太大
	Default_size    int16 `yaml:"default_size"`    // default warehouse size
	Expansion_costs int16 `yaml:"expansion_costs"` // the amount required for expanding size 1
}

type loginConfig struct {
	Model             string `yaml:"model"` // login model
	Other_information struct {
		Username string `yaml:"username"` // login username
		Password string `yaml:"password"` // login password
	} `yaml:"other_information"`
}

func main() {
	fmt.Println("run start!")
	data, err := ioutil.ReadFile("document/config.yaml")
	if err != nil {
		// operation.go file NewStart function.
		panic("Error reading file!")
	}
	var config ConfigFile
	//fmt.Println(string(data))
	err = yaml.Unmarshal(data, &config)
	//fmt.Println(config)
	if err != nil {
		panic("Error converting yaml format!")
	}
	fmt.Println(config)
	// check config product information file exists.
	_, err = os.Stat(config.Config.Goods.Import_file)
	if os.IsNotExist(err) {

		str := fmt.Sprintf("goods information file not exists! file path is %s\n", config.Config.Goods.Import_file)
		panic(str)
	}
}
