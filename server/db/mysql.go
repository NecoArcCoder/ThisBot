package db

import (
	"database/sql"
	"fmt"
	"log"
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

	err = db.Ping()
	if err != nil {
		defer db.Close()
		db = nil
	}

	return db
}
