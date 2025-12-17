package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func del_log(ary []string) {
	if len(ary) < 3 {
		fmt.Println("[💀] Usage: log del [log_id/all], please enter right command")
		return
	}

	if ary[2] == "all" {
		sqlStr := "truncate table logs"
		_, err := db1.Exec(common.Db, sqlStr)
		if err != nil {
			fmt.Println("[💀] Failed to delete logs")
			return
		}
		fmt.Println("[✅] Delete logs successfully")
		return
	}

	id, err := strconv.ParseInt(ary[2], 10, 64)
	if err != nil {
		fmt.Println("[💀] Please enter right task ID")
		return
	}
	sqlStr := "delete from logs where id=?"
	row, err := db1.Exec(common.Db, sqlStr, id)
	if err != nil {
		fmt.Printf("[💀] Failed to delete log[%d]\n", id)
		return
	} else if row == 0 {
		fmt.Printf("[💀] log[%d] doesn't exist\n", id)
		return
	}

	fmt.Printf("[✅] log[%d] deleted okay\n", id)
}

func export_logs(ary []string) {
	sqlStr := "select id,account_id,task_id,client_id,action,message,status,created_at,ip from logs"
	rows, err := db1.QueryRows(common.Db, sqlStr)
	if err != nil {
		fmt.Println("[💀] Error in exporting logs")
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
		if !utils.IsSameSuffix(ary[2], "csv") {
			fmtTime += "_tasks.csv"
		}
	}

	file, err := os.Create(fmtTime)
	if err != nil {
		fmt.Println("[💀] Error in exporting logs")
		return
	}
	defer file.Close()

	// Create CVS writter
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers to csv file
	err = writer.Write([]string{"ID", "AccountID", "TaskID", "BotID", "Action", "Message", "Status", "Create Time", "IP"})
	if err != nil {
		fmt.Println("[💀] Error in exporting logs")
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

	fmt.Println("[✅] Export to " + fmtTime)
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
		fmt.Println("[✅] Failed to list logs")
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
