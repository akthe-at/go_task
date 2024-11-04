package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type CRUD interface {
	Create(db *sql.DB) error
	Read(db *sql.DB) (interface{}, error)
	ReadAll(db *sql.DB) ([]interface{}, error)
	Update(db *sql.DB) (sql.Result, error)
	Delete(db *sql.DB) error
	DeleteMultiple(db *sql.DB, ID ...int) error
}

/*
TaskTable struct to represent the Task Table in the database.
 1. Contains methods to create, read, update, and delete tasks
 2. Uses the Task struct to represent a singular task.
*/
type TaskTable struct {
	Task Task
}

type Task struct {
	ID             int
	Title          string
	Priority       string
	Status         string
	Archived       bool
	UpdateArchived bool
	CreatedAt      time.Time
	LastModified   time.Time
	DueDate        time.Time
	Notes          []Note
}

func (t *Task) Create(db *sql.DB) error {
	return nil
}

func (t *Task) Read(db *sql.DB) (interface{}, error) {
	return nil, nil
}

func (t *Task) ReadAll(db *sql.DB) ([]interface{}, error) {
	return nil, nil
}

func (t *Task) Update(db *sql.DB) (sql.Result, error) {
	return nil, nil
}

func (t *Task) UpdateMultiple(db *sql.DB, ID ...int) error {
	return nil
}

func (t *Task) Delete(db *sql.DB) error {
	if t.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", t.ID)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM task_notes where task_id = ?`, t.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete task notes: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM tasks WHERE id = ?`, t.ID)
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

func (t *Task) DeleteMultiple(db *sql.DB, ID ...int) error {
	return nil
}

/*
Create inserts a new task into the database.
 1. Sets the created_at and last_mod fields to the current time.
 2. If the due_date is not set, it defaults to 7 days from now.
 3. Returns an error if any occurs during the insertion.
*/
func (t *TaskTable) Create(db *sql.DB) error {
	now := time.Now()
	t.Task.CreatedAt = now
	t.Task.LastModified = now
	// TODO: This is just a filler for DueDate, should be set by user...but I don't
	// want to worry about user input yet while testing this out.
	if t.Task.DueDate.IsZero() {
		t.Task.DueDate = now.AddDate(0, 0, 7)
	}
	query := `
    INSERT INTO tasks (title, priority, status, archived, created_at, last_mod, due_date)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query, t.Task.Title, t.Task.Priority, t.Task.Status, t.Task.Archived, t.Task.CreatedAt, t.Task.LastModified, t.Task.DueDate)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskTable) Read(db *sql.DB) (Task, error) {
	var task Task
	query := `SELECT id, title, priority, status, archived, created_at, last_mod, due_date FROM tasks WHERE id = ?`
	err := db.QueryRow(query, t.Task.ID).Scan(&task.ID, &task.Title, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
	if err != nil {
		return task, err
	}

	task.Notes, err = getNotes(db, t.Task.ID, "task_notes")
	if err != nil {
		return task, err
	}

	return task, nil
}

func (t *TaskTable) ReadAll(db *sql.DB) ([]Task, error) {
	var tasks []Task
	query := `SELECT id, title, priority, status, archived, created_at, last_mod, due_date FROM tasks`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var task Task

		err := rows.Scan(&task.ID, &task.Title, &task.Priority, &task.Status, &task.Archived, &task.CreatedAt, &task.LastModified, &task.DueDate)
		if err != nil {
			return nil, err
		}

		task.Notes, err = getNotes(db, task.ID, "task_notes")
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *TaskTable) Update(db *sql.DB) (sql.Result, error) {
	t.Task.LastModified = time.Now()

	queryParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if t.Task.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, t.Task.Title)
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
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(queryParts, ", "), argCounter)
	args = append(args, t.Task.ID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}
	return result, nil
}

func (t *TaskTable) Delete(db *sql.DB) error {
	if t.Task.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", t.Task.ID)
	}

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

func (t *TaskTable) DeleteMultiple(db *sql.DB, ids []int) error {
	return deleteMultiple(db, "tasks", ids)
}

/***************************************************** Area ******************************************************/

type Area struct {
	ID             int
	Title          string
	Type           string
	Tasks          []Task
	Status         string
	Archived       bool
	UpdateArchived bool
	CreatedAt      time.Time
	LastModified   time.Time
	DueDate        time.Time
	Notes          []Note
}

type Note struct {
	ID    int
	Title string
	Path  string
}

type TaskNote struct {
	NoteID int
	TaskID int
}

type AreaNote struct {
	NoteID int
	AreaID int
}

type AreaTable struct {
	Area Area
}

func (a *AreaTable) Create(db *sql.DB) error {
	query := `
		INSERT INTO areas (title, type, deadline, status, archived)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, a.Area.Title, a.Area.Type, a.Area.DueDate, a.Area.Status, a.Area.Archived)
	if err != nil {
		return fmt.Errorf("Failed to create Area/Project in DB: %w", err)
	}

	return nil
}

func (a *AreaTable) Read(db *sql.DB) (Area, error) {
	var area Area
	query := `SELECT id, title, type, deadline, status, archived FROM areas WHERE id = ?`
	err := db.QueryRow(query, a.Area.ID).Scan(&area.ID, &area.Title, &area.Type, &area.DueDate, &area.Status, &area.Archived)
	if err != nil {
		return area, err
	}

	area.Notes, err = getNotes(db, a.Area.ID, "area_notes")
	if err != nil {
		return area, err
	}

	return area, nil
}

/*
ReadAll retrieves all areas from the database.
 1. Returns a slice of areas and an error if any occurs.
 2. Uses a loop to scan each row into an Area struct.
 3. Calls getNotes to retrieve associated notes for each area.
 4. Returns the slice of areas and nil if successful.
*/
func (a *AreaTable) ReadAll(db *sql.DB) ([]Area, error) {
	var areas []Area
	query := `SELECT id, title, type, deadline, status, archived FROM areas`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var area Area
		err := rows.Scan(&area.ID, &area.Title, &area.Type, &area.DueDate, &area.Status, &area.Archived)
		if err != nil {
			return nil, err
		}

		area.Notes, err = getNotes(db, area.ID, "area_notes")
		if err != nil {
			return nil, err
		}
		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

func (a *AreaTable) Update(db *sql.DB) (results sql.Result, err error) {
	queryParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if a.Area.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, a.Area.Title)
		argCounter++
	}
	if a.Area.Type != "" {
		queryParts = append(queryParts, fmt.Sprintf("type = $%d", argCounter))
		args = append(args, a.Area.Type)
		argCounter++
	}
	if a.Area.Status != "" {
		queryParts = append(queryParts, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, a.Area.Status)
		argCounter++
	}
	if a.Area.UpdateArchived != false {
		queryParts = append(queryParts, fmt.Sprintf("archived = $%d", argCounter))
		args = append(args, a.Area.Archived)
		argCounter++
	}

	if !a.Area.DueDate.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("deadline = $%d", argCounter))
		args = append(args, a.Area.DueDate)
		argCounter++
	}

	if len(queryParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	query := fmt.Sprintf("UPDATE areas SET %s WHERE id = $%d", strings.Join(queryParts, ", "), argCounter)
	args = append(args, a.Area.ID)

	results, err = db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}
	return results, nil
}

func (a *AreaTable) Delete(db *sql.DB) error {
	if a.Area.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", a.Area.ID)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	_, err = tx.Exec(`DELETE FROM area_notes where area_id = ?`, a.Area.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete area notes: %w", err)
	}
	_, err = tx.Exec(`DELETE FROM areas WHERE id = ?`, a.Area.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete area: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (a *AreaTable) DeleteMultiple(db *sql.DB, ids []int) error {
	return deleteMultiple(db, "areas", ids)
}

/******************** Helper Functions ********************/

/*
deleteMultiple deletes multiple rows from a table based on the provided IDs.
This function is called by the DeleteMultiple methods of the TaskTable and AreaTable structs.
*/
func deleteMultiple(db *sql.DB, tableName string, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided for deletion")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", tableName, strings.Trim(strings.Repeat("?,", len(ids)), ","))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	return nil
}

/*
getNotes retrieves notes associated with a specific task or area.
 1. Takes a database connection, an ID, and a table name as parameters.
 2. Returns a slice of notes and an error if any occurs.
*/
func getNotes(db *sql.DB, id int, table string) ([]Note, error) {
	notesQuery := fmt.Sprintf(`
        SELECT notes.id, notes.title, notes.path
        FROM notes
        JOIN %s ON notes.id = %s.note_id
        WHERE %s.%s_id = ?
    `, table, table, table, table[:len(table)-6]) // table[:len(table)-6] removes the "_notes" suffix

	rows, err := db.Query(notesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
