package main

type ConfigFile struct {
	Config struct {
		Database struct { // database information
			Type     string `yaml:"type"`
			Database string `yaml:"database"`
			Address  string `yaml:"address"`
			Port     string `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"database"`
	} `yaml:"config"`
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
}

type warehouseConfig struct {
	// 此为 默认仓库大小 和 扩容1点仓库所需金额 不需要太大
	Default_size    int16 `yaml:"default_size"`    // default warehouse size
	expansion_costs int16 `yaml:"expansion_costs"` // the amount required for expanding size 1
}

type loginConfig struct {
	Model             string `yaml:"model"` // login model
	Other_information struct {
		Username string `yaml:"username"` // login username
		Password string `yaml:"password"` // login password
	} `yaml:"other_information"`
}
