package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Area struct {
	ID             int
	Title          string
	Tasks          []Task
	Status         StatusType
	Archived       bool
	UpdateArchived bool
	CreatedAt      time.Time
	LastModified   time.Time
	DueDate        time.Time
	Notes          []Note
}

// Create inserts a new row into the areas table.
func (a *Area) Create(db *sql.DB) error {
	query := `
		INSERT INTO areas (title, status, archived, created_at, last_mod, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, a.Title, a.Status, a.Archived, a.CreatedAt, a.LastModified, a.DueDate)
	if err != nil {
		return fmt.Errorf("Failed to create Area/Project in DB: %w", err)
	}

	return nil
}

// Read retrieves a single row from the areas table based on the provided ID.
func (a *Area) Read(db *sql.DB) error {
	query := `SELECT id, title, deadline, status, archived FROM areas WHERE id = ?`
	err := db.QueryRow(query, a.ID).Scan(&a.ID, &a.Title, &a.DueDate, &a.Status, &a.Archived)
	if err != nil {
		return fmt.Errorf("AreaRead: failed to read area: %w", err)
	}

	// Fetch any associated notes
	rows, err := db.Query(`
SELECT notes.id, notes.title, notes.path
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_area_id = ?
and bridge_notes.parent_cat = 2
		`, a.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch area notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path); err != nil {
			return rows.Err()
		}
		a.Notes = append(a.Notes, note)
	}

	return nil
}

/*
ReadAll retrieves all areas from the database.
 1. Returns a slice of areas and an error if any occurs.
 2. Uses a loop to scan each row into an Area struct.
 3. Calls getNotes to retrieve associated notes for each area.
 4. Returns the slice of areas and nil if successful.
*/
func (a *Area) ReadAll(db *sql.DB) ([]Area, error) {
	var areas []Area
	query := `SELECT id, title, deadline, status, archived FROM areas`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var area Area
		err := rows.Scan(&area.ID, &area.Title, &area.DueDate, &area.Status, &area.Archived)
		if err != nil {
			return nil, err
		}

		// Fetch related notes for the area
		notesQuery := `
		SELECT notes.id, notes.title, notes.path 
		FROM notes 
		INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
		WHERE parent_cat = 2 AND parent_area_id = ?`
		notesRows, err := db.Query(notesQuery, area.ID)
		if err != nil {
			return nil, err
		}
		defer notesRows.Close()

		for notesRows.Next() {
			var note Note
			err := notesRows.Scan(&note.ID, &note.Title, &note.Path)
			if err != nil {
				return nil, err
			}
			area.Notes = append(area.Notes, note)
		}

		if err = notesRows.Err(); err != nil {
			return nil, err
		}

		areas = append(areas, area)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return areas, nil
}

// Update updates a single row in the areas table based on the provided ID.
func (a *Area) Update(db *sql.DB) (results sql.Result, err error) {
	queryParts := []string{}
	args := []interface{}{}
	argCounter := 1

	if a.Title != "" {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, a.Title)
		argCounter++
	}
	if a.Status != "" {
		queryParts = append(queryParts, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, a.Status)
		argCounter++
	}
	if a.UpdateArchived != false {
		queryParts = append(queryParts, fmt.Sprintf("archived = $%d", argCounter))
		args = append(args, a.Archived)
		argCounter++
	}

	if !a.DueDate.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("deadline = $%d", argCounter))
		args = append(args, a.DueDate)
		argCounter++
	}

	if len(queryParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	query := fmt.Sprintf("UPDATE areas SET %s WHERE id = $%d", strings.Join(queryParts, ", "), argCounter)
	args = append(args, a.ID)

	results, err = db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}
	return results, nil
}

// Delete removes a single row from the areas table based on the provided ID.
func (a *Area) Delete(db *sql.DB) error {
	if a.ID == 0 {
		return fmt.Errorf("invalid task ID: %d", a.ID)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`
		DELETE FROM notes
		where notes.id IN (
		SELECT bridge_notes.note_id
		FROM bridge_notes
		WHERE parent_cat = 2 AND parent_area_id = ?
		)`, a.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete related area notes: %w", err)
	}
	_, err = tx.Exec(`DELETE FROM areas WHERE id = ?`, a.ID)
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

// DeleteMultiple deletes multiple rows in the areas table based on the provided IDs.
func (a *Area) DeleteMultiple(db *sql.DB, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	targetIDs := strings.Trim(strings.Repeat("?,", len(ids)), ",")
	// Construct the query with placeholders for the IDs - Allows for variable number of IDs
	query := fmt.Sprintf("DELETE FROM areas WHERE id IN (%s)", targetIDs)
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete areas: %w", err)
	}

	// Construct the query to delete associated notes
	notesQuery := fmt.Sprintf(`
		DELETE FROM notes
		WHERE id IN (
			SELECT note_id
			FROM bridge_notes
			WHERE parent_cat = 2 AND parent_area_id IN (%s)
		)
	`, targetIDs)

	_, err = tx.Exec(notesQuery, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete area notes: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
