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
	// TODO: This is just a filler for DueDate, should be set by user.
	if t.Task.DueDate.IsZero() {
		t.Task.DueDate = now.AddDate(0, 0, 7)
	}
	query := `
    INSERT INTO tasks (title, description, priority, status, archived, created_at, last_mod, due_date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query, t.Task.Title, t.Task.Description, t.Task.Priority, t.Task.Status, t.Task.Archived, t.Task.CreatedAt, t.Task.LastModified, t.Task.DueDate)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskTable) Read(db *sql.DB) (data.Task, error) {
	var task data.Task
	query := `SELECT id, title, description, priority, status, archived, created_at, last_mod, due_date FROM tasks WHERE id = ?`
	err := db.QueryRow(query, t.Task.ID).Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
	if err != nil {
		return task, err
	}

	notesQuery := `
		SELECT notes.id, notes.title, notes.path
		FROM notes
		JOIN task_notes ON notes.id = task_notes.note_id
		WHERE task_notes.task_id = ?
	`
	rows, err := db.Query(notesQuery, t.Task.ID)
	if err != nil {
		return task, err
	}
	defer rows.Close()

	for rows.Next() {
		var note data.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
			return task, err
		}
		task.Notes = append(task.Notes, note)
	}

	if err := rows.Err(); err != nil {
		return task, err
	}

	return task, nil
}

func (t *TaskTable) ReadAll(db *sql.DB) ([]data.Task, error) {
	var tasks []data.Task
	query := `SELECT id, title, description, priority, status, archived, created_at, last_mod, due_date FROM tasks`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var task data.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
		if err != nil {
			return nil, err
		}

		notesQuery := `
			SELECT notes.id, notes.title, notes.path
			FROM notes
			JOIN task_notes ON notes.id = task_notes.note_id
			WHERE task_notes.task_id = ?
		`
		noteRows, err := db.Query(notesQuery, task.ID)
		if err != nil {
			return nil, err
		}
		defer noteRows.Close()

		for noteRows.Next() {
			var note data.Note
			if err := noteRows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
				return nil, err
			}
			task.Notes = append(task.Notes, note)
		}

		if err := noteRows.Err(); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
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
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM task_notes where task_id = ?`, t.Task.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete task notes: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM tasks WHERE id = ?`, t.Task.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete task: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (t *TaskTable) DeleteMultiple(db *sql.DB, ID ...int) error {
	if len(ID) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := "DELETE FROM tasks WHERE id IN ("
	args := make([]interface{}, len(ID))

	for i, id := range ID {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	notesQuery := strings.Replace(query, "tasks", "task_notes", 1)
	notesQuery = strings.Replace(notesQuery, "id", "task_id", 1)
	_, err = tx.Exec(notesQuery, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete task notes: %w", err)
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete tasks: %w", err)
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

type AreaTable struct {
	Area data.Area
}

func (a *AreaTable) Create(db *sql.DB) error {
	query := `
		INSERT INTO areas (title, type, deadline, status, archived)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, a.Area.Title, a.Area.Type, a.Area.Deadline, a.Area.Status, a.Area.Archived)
	if err != nil {
		return fmt.Errorf("Failed to create Area/Project in DB: %w", err)
	}

	return nil
}

func (a *AreaTable) Read(db *sql.DB) (data.Area, error) {
	var area data.Area
	query := `SELECT id, title, type, deadline, status, archived FROM areas WHERE id = ?`
	err := db.QueryRow(query, a.Area.ID).Scan(&area.ID, &area.Title, &area.Type, &area.Deadline, &area.Status, &area.Archived)
	if err != nil {
		return area, err
	}
	notesQuery := `
		SELECT notes.id, notes.title, notes.path
		FROM notes
		JOIN area_notes ON notes.id = area_notes.note_id
		WHERE area_notes.area_id = ?
	`
	rows, err := db.Query(notesQuery, area.ID)
	if err != nil {
		return area, err
	}
	defer rows.Close()
	for rows.Next() {
		var note data.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
			return area, err
		}
		area.Notes = append(area.Notes, note)
	}
	if err := rows.Err(); err != nil {
		return area, err
	}
	return area, nil
}

func (a *AreaTable) ReadAll(db *sql.DB) ([]data.Area, error) {
	var areas []data.Area
	query := `SELECT id, title, type, deadline, status, archived FROM areas`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var area data.Area
		err := rows.Scan(&area.ID, &area.Title, &area.Type, &area.Deadline, &area.Status, &area.Archived)
		if err != nil {
			return nil, err
		}
		notesQuery := `
			SELECT notes.id, notes.title, notes.path
			FROM notes
			JOIN area_notes ON notes.id = area_notes.note_id
			WHERE area_notes.area_id = ?
		`
		noteRows, err := db.Query(notesQuery, area.ID)
		if err != nil {
			return nil, err
		}
		defer noteRows.Close()
		for noteRows.Next() {
			var note data.Note
			if err := noteRows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
				return nil, err
			}
			area.Notes = append(area.Notes, note)
		}
		if err := noteRows.Err(); err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	return areas, nil
}
