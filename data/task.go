package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type (
	PriorityType string
	StatusType   string
)

const (
	PriorityTypeLow    PriorityType = "low"
	PriorityTypeMedium PriorityType = "medium"
	PriorityTypeHigh   PriorityType = "high"
	PriorityTypeUrgent PriorityType = "urgent"
	StatusToDo         StatusType   = "todo"
	StatusPlanning     StatusType   = "planning"
	StatusDoing        StatusType   = "doing"
	StatusDone         StatusType   = "done"
)

type CRUD interface {
	Create(db *sql.DB) error
	Read(db *sql.DB) ([]interface{}, error)
	// ReadByID(db *sql.DB, id int) (interface{}, error)
	ReadAll(db *sql.DB) ([]interface{}, error)
	Update(db *sql.DB) (sql.Result, error)
	Delete(db *sql.DB) error
	DeleteMultiple(db *sql.DB, ID ...int) error
	Query(db *sql.DB, query string) ([]interface{}, error)
}

type Task struct {
	ID             int
	Title          string
	Priority       PriorityType
	Status         StatusType
	Archived       bool
	UpdateArchived bool
	CreatedAt      time.Time
	LastModified   time.Time
	DueDate        time.Time
	TaskAge        float32
	Notes          []Note
	NoteTitles     string
	Area           *Area
}

// Create creates a new task in the database.
func (t *Task) Create(db *sql.DB) error {
	now := time.Now()
	t.CreatedAt = now
	t.LastModified = now
	query := `
    INSERT INTO tasks (title, priority, status, archived, created_at, last_mod, due_date)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query, t.Title, t.Priority, t.Status, t.Archived, t.CreatedAt, t.LastModified, t.DueDate)
	if err != nil {
		return err
	}
	return nil
}

// Read retrieves a task from the database based on the provided ID.
func (t *Task) Read(db *sql.DB) error {
	if t.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", t.ID)
	}

	query := `SELECT id, title, priority, status, archived, created_at, last_mod, ROUND((julianday('now') - julianday(tasks.created_at)),2) AS age_in_days, due_date, FROM tasks WHERE id = ?`
	row := db.QueryRow(query, t.ID)
	err := row.Scan(&t.ID, &t.Title, &t.Priority, &t.Status, &t.Archived, &t.CreatedAt, &t.LastModified, &t.TaskAge, &t.DueDate)
	if err != nil {
		return fmt.Errorf("failed to read task: %w", err)
	}

	// Fetch any associated notes
	rows, err := db.Query(`
SELECT notes.id, notes.title, notes.path
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id = ? 
AND bridge_notes.parent_cat = 1
		`, t.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch area notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
			return rows.Err()
		}
		t.Notes = append(t.Notes, note)
	}
	return nil
}

// ReadAll retrieves all tasks from the database.
func (t *Task) ReadAll(db *sql.DB) ([]Task, error) {
	var tasks []Task
	query := `
	SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)),2) AS age_in_days,
		IFNULL(GROUP_CONCAT(notes.title, ', '), '') as note_titles
	FROM tasks
	LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
	LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
	GROUP BY tasks.id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task

		err := rows.Scan(&task.ID, &task.Title, &task.Priority, &task.Status, &task.Archived, &task.TaskAge, &task.NoteTitles)
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

// Update updates a task in the database based on the provided ID.
func (t *Task) Update(db *sql.DB) (sql.Result, error) {
	t.LastModified = time.Now()

	queryParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if t.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, t.Title)
		argCounter++
	}

	if t.Priority != "" {
		queryParts = append(queryParts, fmt.Sprintf("priority = $%d", argCounter))
		args = append(args, t.Priority)
		argCounter++
	}

	if t.Status != "" {
		queryParts = append(queryParts, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, t.Status)
		argCounter++
	}

	if t.UpdateArchived != false {
		queryParts = append(queryParts, fmt.Sprintf("archived = $%d", argCounter))
		args = append(args, t.Archived)
		argCounter++
	}

	if !t.DueDate.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("due_date = $%d", argCounter))
		args = append(args, t.DueDate)
		argCounter++
	}

	queryParts = append(queryParts, fmt.Sprintf("last_mod = $%d", argCounter))
	args = append(args, t.LastModified)
	argCounter++

	if len(queryParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(queryParts, ", "), argCounter)
	args = append(args, t.ID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}
	return result, nil
}

// // UpdateMultiple updates multiple rows in a table based on the provided IDs.
// func (t *Task) UpdateMultiple(db *sql.DB, ID ...int) error {
// 	return error.
// }

func (t *Task) Delete(db *sql.DB) error {
	if t.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", t.ID)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`
		DELETE FROM notes
		WHERE id IN (
			SELECT note_id
			FROM bridge_notes
			WHERE parent_cat = 1 AND parent_task_id = ?
		)
		`, t.ID)
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

// DeleteMultiple deletes multiple rows from a table based on the provided IDs.
func (t *Task) DeleteMultiple(db *sql.DB, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	targetIDs := strings.Trim(strings.Repeat("?,", len(ids)), ",")
	// Construct the query with placeholders for the IDs - Allows for variable number of IDs
	query := fmt.Sprintf("DELETE FROM tasks WHERE id IN (%s)", targetIDs)
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	// Construct the query to delete associated notes
	notesQuery := fmt.Sprintf(`
		DELETE FROM notes
		where notes.id IN (
		SELECT bridge_notes.note_id
		FROM bridge_notes
		WHERE parent_cat = 1 AND parent_task = ? IN (%s)
		)
	`, targetIDs)

	_, err = tx.Exec(notesQuery, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete task notes: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (t *Task) Query(db *sql.DB, query string) error {
	return QueryAndPrint(db, query)
}

func QueryAndPrint(db *sql.DB, query string) error {
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		for i, col := range columns {
			fmt.Printf("%s: %v\n", col, values[i])
		}
		fmt.Println("-----")
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}
