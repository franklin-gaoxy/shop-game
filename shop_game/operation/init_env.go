package operation

import (
	"shop_game/shop_game/database"
	"shop_game/shop_game/tools"
)

func initEnvironment(configfilename string) {
	var config tools.ConfigFile
	config = tools.ReadConfig(configfilename)
	//NewDB(&DB, config)

	databaseOper, err := database.IsSupported(config.Config.Database.Type)
	if err != nil {
		panic("init database error!")
	}
	// conn
	databaseOper.Conn(config)

	// execute code
	databaseOper.Init()
}

var userInfo tools.UserInformation
