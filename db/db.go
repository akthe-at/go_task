package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/akthe-at/go_task/data"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// ConnectDB opens a connection to a SQLite database.
// It returns a pointer to the sql.DB object and an error if any occurs.
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:new_demo.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

// isDatabaseSetup checks if the database has been set up.
func IsDatabaseSetup(db *sql.DB) bool {
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

// SetupDB Setup the Initial DB Schema
// 1. Creates the areas and tasks tables if they do not exist
// 2. Returns an error if any occurs
// 3. Uses transactions for safety
// 4. Uses prepared statements for better performance
func SetupDB(db *sql.DB) error {
	initialQueries := `
		CREATE TABLE IF NOT EXISTS areas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			type TEXT,
			deadline TEXT,
			status INTEGER,
			archived BOOLEAN
	);
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			priority TEXT,
			status INTEGER,
			archived BOOLEAN,
			created_at TEXT,
			last_mod TEXT,
			due_date TEXT
	);
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			path TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS task_notes (
			note_id INTEGER,
			task_id INTEGER,
			FOREIGN KEY(note_id) REFERENCES notes(id),
			FOREIGN KEY(task_id) REFERENCES tasks(id)
		);
		CREATE TABLE IF NOT EXISTS area_notes (
			note_id INTEGER,
			area_id INTEGER,
			FOREIGN KEY(note_id) REFERENCES notes(id),
			FOREIGN KEY(area_id) REFERENCES areas(id)
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

func UpdateNotePath(db *sql.DB, noteID int, newPath string) error {
	query := "UPDATE notes SET path = ? WHERE id = ?"
	_, err := db.Exec(query, newPath, noteID)
	return err
}
