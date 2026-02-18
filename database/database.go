package database

import (
	"log"

	"apiTrackingSystem/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func MustOpen(cfg config.Config) *sqlx.DB {
	db, err := sqlx.Open("mysql", cfg.DBDSN)
	if err != nil {
		log.Fatal("db open:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}

	return db
}

