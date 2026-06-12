package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ConnDB struct {
	db *sql.DB
}

func NewConnDB(dsn string) (*ConnDB, error) {
	if len(dsn) == 0 {
		dsn = "./data/forum.db"
	}

	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		log.Println("Connection to sqlite failed")

		return nil, err
	}

	log.Println("sqlite3 connected successfully")

	additional_options := `
		PRAGMA foreign_keys = ON;
		PRAGMA journal_mode = WAL;
		PRAGMA busy_timeout = 5000;
		PRAGMA synchronous = NORMAL;
	`

	if _, err := db.Exec(additional_options); err != nil {
		return nil, err
	}

	return &ConnDB{db: db}, nil
}

func (c *ConnDB) Close() error {
	return c.db.Close()
}

func (c *ConnDB) GetDB() *sql.DB {
	return c.db
}
