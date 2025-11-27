package db

import (
	"database/sql"
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

	err = db.Ping()
	if err != nil {
		defer db.Close()
		db = nil
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
