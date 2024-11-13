package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// ConnectDB opens a connection to a SQLite database.
// It returns a pointer to the sql.DB object and an error if any occurs.
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:new_demo.db")
	if err != nil {
		return nil, fmt.Errorf("invalid sql.Open() arguments: %w", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("failed to ping database: ", err)
	}
	return db, nil
}

// FileExists This function checks if a file exists, if it does returns true.
// Otherwise it returns false.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

/*
IsSetup checks if the database has been set up.
 1. Returns false if the tasks table does not exist
 2. Returns true if the tasks table exists
*/
func IsSetup(db *sql.DB) bool {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='tasks'`
	var name string
	err := db.QueryRow(query).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
	}
	return true
}

/*
SetupDB Setup the Initial DB Schema
 1. Creates the areas and tasks tables if they do not exist
 2. Returns an error if any occurs
 3. Uses transactions for safety
 4. Uses prepared statements for better performance
*/
func SetupDB(db *sql.DB) error {
	initialQueries := `
PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS areas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			status TEXT,
			archived BOOLEAN
			created_at DATETIME,
			last_mod DATETIME,
			due_date DATETIME
	);
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			priority TEXT,
			status TEXT,
			archived BOOLEAN,
			created_at DATETIME,
			last_mod DATETIME,
			due_date DATETIME
	area_id INTEGER,
	FOREIGN KEY(area_id) REFERENCES areas(id) ON DELETE SET NULL ON UPDATE CASCADE
	);
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			path TEXT NOT NULL
		);
		CREATE TABLE bridge_notes (
			note_id INTEGER,
			parent_cat INTEGER,
			parent_task_id INTEGER,
			parent_area_id INTEGER,
			FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE ON UPDATE CASCADE,
			CHECK (parent_cat IN (1, 2)),
			FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
			FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE ON UPDATE CASCADE
		);
`
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(initialQueries)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create table: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ResetDB drops all tables and recreates them.
func ResetDB(db *sql.DB) error {
	queries := `
		DROP TABLE IF EXISTS areas;
		DROP TABLE IF EXISTS tasks;
		DROP TABLE IF EXISTS notes;
		DROP TABLE IF EXISTS bridge_notes;
	`
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	_, err = tx.Exec(queries)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	err = SetupDB(db)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}

	return nil
}

func UpdateNotePath(db *sql.DB, noteID int, newPath string) error {
	query := "UPDATE notes SET path = ? WHERE id = ?"
	_, err := db.Exec(query, newPath, noteID)
	return err
}
