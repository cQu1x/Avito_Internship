package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(DSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
