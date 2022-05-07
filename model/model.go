package model

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

var db *sql.DB

func Setup () {
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(localhost:3306)/test")
	if err != nil {
		log.Fatalf("model.Setup err %v", err)
	}
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
}

func CloseDb () {
	defer db.Close()
}


func ParseStringToTime (timestamp string) (time.Time, error) {
	location , _ := time.LoadLocation("Local")
	return time.ParseInLocation("2006-01-02 15:04:05",  timestamp, location)
}

func NewDB () *sql.DB {
	return db
}

