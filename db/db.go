package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DBManager struct {
	Db            *sql.DB
	IsInitialized bool
}

var dbInstance = new()

// GetDB return DB instance
func GetDB() DBManager {
	return dbInstance
}

// NewDB inicializa POSTGRESQL
func new() DBManager {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		panic(err)
	}
	return DBManager{db, true}
}
