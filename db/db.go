package db

import (
	"database/sql"
	"os"
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
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
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
