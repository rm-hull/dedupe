package internal

import (
	"database/sql"
	"fmt"
)

func Connect(username string, password string, host string) (*sql.DB, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/dedupe?sslmode=disable", username, password, host)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	return db, nil
}
