package main

import (
	"ThisBot/common"
	"ThisBot/config"
	"ThisBot/utils"
	"log"
	"os"
	"path"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--init-config" {
		err := config.GenerateDefaultConfig(common.ConfigDefaultFileName)
		if err != nil {
			log.Fatal("config.GenerateDefaultConfig failed")
			return
		}
		log.Println("Generate default configure file successfully, please check it and restart the server")
		return
	}

	// Check the configure file exists or not
	config_path, _ := os.Getwd()
	config_path = path.Join(config_path, common.ConfigDefaultFileName)
	if !utils.FileExist(config_path) {
		log.Fatal("Can't find configure file, please add --init-config option to generate it first")
		return
	}
	// Load config
	pCfg := config.LoadConfig(config_path)
	if pCfg == nil {
		log.Fatal("Failed to load configure file")
		return
	}
	common.Cfg = *pCfg

	// Initialize all
	config.Init(&common.Cfg)

	// Running the server
	Server()
}
