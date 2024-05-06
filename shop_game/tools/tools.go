package tools

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

const (
	End    string = "\033[0m"
	Red    string = "\033[31m"
	Green  string = "\033[32m"
	Yellow string = "\033[33m"
)

// 用于捕获用户输入操作
var Oper string

func ReEnter() {
	log.Println(Red + "No such operation, please re-enter" + End)
}

func EncryptPassword(password string) string {
	// 将密码字符串转换为字节数组
	passwordBytes := []byte(password)
	// 计算MD5哈希值
	hash := md5.Sum(passwordBytes)
	// 将哈希值转换为16进制字符串
	hashedPassword := hex.EncodeToString(hash[:])
	return hashedPassword
}

func ReadGoods(path string) AllGoods {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic("read goods file error!")
	}

	var goods AllGoods
	err = yaml.Unmarshal(data, &goods)
	if err != nil {
		log.Fatalf("[tools.go:ReadGoods()]: Failed to unmarshal YAML: %v", err)
	}
	return goods
}

func ReadConfig(configfilename string) ConfigFile {
	// load config.
	data, err := ioutil.ReadFile(configfilename)
	if err != nil {
		// operation.go file NewStart function.
		panic("[operation.go:readConfig]: Error reading file!")
	}
	var config ConfigFile
	err = yaml.Unmarshal(data, &config)
	//fmt.Println(config)
	if err != nil {
		panic("[operation.go:readConfig]: Error converting yaml format!")
	}

	// check config product information file exists.
	_, err = os.Stat(config.Config.Goods.Import_file)
	if os.IsNotExist(err) {
		str := fmt.Sprintf("[operation.go:readConfig]: goods information file not exists! file path is %s\n", config.Config.Goods.Import_file)
		panic(str)
	}

	//fmt.Println("配置文件内容:", config)
	log.Println("Configuration file reading completed.")
	return config
}
