package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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

func Migrate(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	m.Up()
	if err != nil {
		return err
	}

	return nil
}

func InsertFileEntryStatement(db *sql.DB) (*sql.Stmt, error) {
	return db.Prepare(
		"INSERT INTO file_entry " +
			"(scan_id, name, size, mode, mod_time, is_dir, hash) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7)")
}

func CreateScan(db *sql.DB, root string) (*int64, error) {

	absolutePath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	var scanId int64
	err = db.QueryRow(
		"INSERT INTO scan "+
			"(hostname, scan_status, root_directory) "+
			"VALUES ($1, $2, $3) RETURNING id", hostname, InProgress.String(), absolutePath).Scan(&scanId)
	if err != nil {
		return nil, err
	}

	return &scanId, nil
}
