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

type TaskTable struct {
	Task data.Task
}

func (t *TaskTable) Create(db *sql.DB) error {
	now := time.Now()
	t.Task.CreatedAt = now
	t.Task.LastModified = now
	if t.Task.DueDate.IsZero() {
		t.Task.DueDate = now.AddDate(0, 0, 7)
	}
	query := `
    INSERT INTO tasks (title, description, priority, status, archived, created_at, last_mod, due_date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query, t.Task.Title, t.Task.Description, t.Task.Priority, t.Task.Status, t.Task.Archived, t.Task.CreatedAt, t.Task.LastModified, t.Task.DueDate)
	return err
}

func (t *TaskTable) Read(db *sql.DB) (data.Task, error) {
	var task data.Task
	query := `SELECT id, title, description, priority, status, archived, created_at, last_mod, due_date FROM tasks WHERE id = ?`
	err := db.QueryRow(query, t.Task.ID).Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
	if err != nil {
		return task, err
	}
	return task, nil
}

func (t *TaskTable) Update(db *sql.DB) error {
	now := time.Now()
	t.Task.LastModified = now

	queryParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if t.Task.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, t.Task.Title)
		argCounter++
	}

	if t.Task.Description != "" {
		queryParts = append(queryParts, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, t.Task.Description)
		argCounter++
	}

	if t.Task.Priority != "" {
		queryParts = append(queryParts, fmt.Sprintf("priority = $%d", argCounter))
		args = append(args, t.Task.Priority)
		argCounter++
	}

	if t.Task.Status != "" {
		queryParts = append(queryParts, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, t.Task.Status)
		argCounter++
	}

	if t.Task.UpdateArchived != false {
		queryParts = append(queryParts, fmt.Sprintf("archived = $%d", argCounter))
		args = append(args, t.Task.Archived)
		argCounter++
	}

	if !t.Task.DueDate.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("due_date = $%d", argCounter))
		args = append(args, t.Task.DueDate)
		argCounter++
	}

	queryParts = append(queryParts, fmt.Sprintf("last_mod = $%d", argCounter))
	args = append(args, t.Task.LastModified)
	argCounter++

	if len(queryParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(queryParts, ", "), argCounter)
	args = append(args, t.Task.ID)

	_, err := db.Exec(query, args...)
	return err
}

func (t *TaskTable) Delete(db *sql.DB) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Exec(query, t.Task.ID)
	return err
}
