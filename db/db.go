package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB opens a connection to a SQLite database.
// It returns a pointer to the sql.DB object and an error if any occurs.
func ConnectDB() (*sql.DB, string, error) {
	var dbPath string
	switch runtime.GOOS {
	case "windows":
		if windowsConfigDir := os.Getenv("LOCALAPPDATA"); windowsConfigDir != "" {
			dbPath = filepath.Join(windowsConfigDir, "go_task")
		} else {
			log.Fatalf("LOCALAPPDATA is not set")
		}
	case "linux":
		if xdgConfig := os.Getenv("XDG_DATA_HOME"); xdgConfig != "" {
			dbPath = filepath.Join(xdgConfig, "go_task")
		} else {
			log.Fatalf("XDG_DATA_HOME is not set")
		}
		err := os.MkdirAll(dbPath, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
	}
	err := os.MkdirAll(dbPath, os.ModePerm)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create directory: %w", err)
	}
	db, err := sql.Open("sqlite3", dbPath+"/taskdb.db")
	if err != nil {
		return nil, "", fmt.Errorf("invalid sql.Open() arguments: %w", err)
	}
	completePath := dbPath + "/taskdb.db"

	err = db.Ping()
	if err != nil {
		return nil, "", fmt.Errorf("failed to ping database: %w", err)
	}
	return db, completePath, nil
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
	queries := []string{
		`PRAGMA foreign_keys=ON;`,
		`CREATE TABLE IF NOT EXISTS areas (
            id INTEGER PRIMARY KEY,
            title TEXT NOT NULL,
            status TEXT,
            archived BOOLEAN NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
            last_mod TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime'))
        );`,
		`CREATE TABLE IF NOT EXISTS task_ids (
            id INTEGER PRIMARY KEY
        );`,
		`CREATE TABLE IF NOT EXISTS area_ids (
            id INTEGER PRIMARY KEY
        );`,
		`CREATE TABLE IF NOT EXISTS prog_proj_ids (
            id INTEGER PRIMARY KEY
        );`,
		`CREATE TABLE IF NOT EXISTS note_ids (
            id INTEGER PRIMARY KEY
        );`,
		`CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY,
            title TEXT NOT NULL,
            priority TEXT,
            status TEXT,
            archived BOOLEAN NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
            last_mod TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
            due_date TEXT,
            area_id INTEGER,
            FOREIGN KEY(area_id) REFERENCES areas(id) ON DELETE SET NULL ON UPDATE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS notes (
            id INTEGER PRIMARY KEY,
            title TEXT NOT NULL,
            path TEXT NOT NULL,
            parent_area_id INTEGER,
            parent_task_id INTEGER
        );`,
		`CREATE TABLE IF NOT EXISTS bridge_notes (
            note_id INTEGER,
            parent_cat INTEGER,
            parent_task_id INTEGER,
            parent_area_id INTEGER,
            FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE ON UPDATE CASCADE,
            CHECK (parent_cat IN (1, 2)),
            FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
            FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE ON UPDATE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS programming_projects (
            id INTEGER PRIMARY KEY,
            path TEXT NOT NULL UNIQUE
        );`,
		`CREATE TABLE IF NOT EXISTS prog_project_links (
            project_id INTEGER,
            parent_cat INTEGER,
            parent_task_id INTEGER,
            parent_area_id INTEGER,
            FOREIGN KEY(project_id) REFERENCES programming_projects(id) ON DELETE CASCADE,
            CHECK (parent_cat IN (1, 2)),
            FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
            FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE
        );`,
		`CREATE TRIGGER IF NOT EXISTS update_last_mod_tasks
        AFTER UPDATE ON tasks
        BEGIN
            UPDATE tasks 
            SET last_mod = datetime(current_timestamp, 'localtime')
            WHERE id = OLD.id;
        END;`,
		`CREATE TRIGGER IF NOT EXISTS update_last_mod_areas
        AFTER UPDATE ON areas
        BEGIN
            UPDATE areas 
            SET last_mod = datetime(current_timestamp, 'localtime')
            WHERE id = OLD.id;
        END;`,
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, query := range queries {
		_, err = tx.Exec(query)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("failed to rollback after error: %v. Original error: %w", rollbackErr, err)
			}
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
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
		DROP TABLE IF EXISTS programming_projects;
		DROP TABLE IF EXISTS prog_project_links;
		DROP TRIGGER IF EXISTS update_last_mod_tasks;
		DROP TRIGGER IF EXISTS update_last_mod_areas;
	`
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	_, err = tx.Exec(queries)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			fmt.Printf("failed to rollback transaction: %v", err)
		}
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
