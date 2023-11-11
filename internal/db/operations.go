package db

import (
	"database/sql"
	"os"
)

func InsertFileEntryStatement(db *sql.DB) (*sql.Stmt, error) {
	return db.Prepare(
		"INSERT INTO file_entry " +
			"(scan_id, name, size, mode, mod_time, is_dir, hash) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7)")
}

func CreateScan(db *sql.DB, root string) (*int64, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	var scanId int64
	err = db.QueryRow(
		"INSERT INTO scan "+
			"(hostname, scan_status, root_directory) "+
			"VALUES ($1, $2, $3) RETURNING id", hostname, InProgress.String(), root).Scan(&scanId)
	if err != nil {
		return nil, err
	}

	return &scanId, nil
}

func UpdateScan(db *sql.DB, scanId int64, filesScanned int, err error) error {
	var errorMessage *string
	if err != nil {
		*errorMessage = err.Error()
	}

	_, err = db.Exec(
		"UPDATE scan "+
			"SET scan_status = $1, finished_at = NOW(), files_scanned = $2, error = $3 "+
			"WHERE id = $4", GetCompletedScanStatus(err).String(), filesScanned, errorMessage, scanId)
	return err
}
