package database

import (
	"crypto/tls"
	"strconv"
	"time"

	"github.com/go-pg/pg/v9"
)

func NewPostgreSQLConnection(user, password, dbName, host, port, sslMode string) (*pg.DB, error) {

	isSslMode, _ := strconv.ParseBool(sslMode)
	options := &pg.Options{
		User:     user,
		Password: password,
		Database: dbName,
		Addr:     host + ":" + port,
	}

	if isSslMode {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: isSslMode}
	}

	dbConn := pg.Connect(options)
	dbConn = dbConn.WithTimeout(90 * time.Second)

	// Test Connection
	_, err := dbConn.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}
