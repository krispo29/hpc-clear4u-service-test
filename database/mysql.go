package database

import (
	"database/sql"
	"fmt"
	"time"
)

func NewMySQLConnection(username, password, host, port, dbName string) (*sql.DB, error) {

	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", username, password, host, port, dbName)
	db, err := sql.Open("mysql", url)
	if err != nil {
		panic(err)
	}

	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// MySQL Test Connection
	_, err = db.Exec("SELECT 1")
	if err != nil {
		panic(err.Error())
	}

	return db, nil
}
