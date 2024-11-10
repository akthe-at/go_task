package data

import (
	"database/sql"
	"fmt"
)

const (
	TaskNoteType NoteType = 1
	AreaNoteType NoteType = 2
)

type NoteTable struct {
	NoteID    int
	NoteTitle string
	NotePath  string
	LinkTitle string
}

type NoteBridge struct {
	noteID        int
	parentCatType int
	parentID      int
}

type NoteType int

type Note struct {
	ID    int
	Title string
	Path  string
	Type  int
}

type NoteCRUD interface {
	Create(db *sql.DB) error
	ReadByID(db *sql.DB, noteID int) (*Note, error)
	ReadByItemID(db *sql.DB, itemID int) ([]Note, error)
	ReadAll(db *sql.DB) ([]Note, error)
	Update(db *sql.DB) error
	Delete(db *sql.DB, noteID int) error
}

// Create creates a new note for a specific task/area/project
func (n *Note) Create(db *sql.DB, noteType NoteType, parentID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	noteQuery := "INSERT INTO notes (title, path, type) VALUES  (?, ?, ?)"
	result, err := tx.Exec(noteQuery, n.Title, n.Path, n.Type)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NoteQueryInsert: %v", err)
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	// TODO: Create this in the database module...
	query := "INSERT INTO bridge_notes (note_id, cat_type, parentID) VALUES (?, ?, ?)"
	_, err = db.Exec(query, noteID, noteType, parentID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("task_noteInsert: %v", err)
	}
	return tx.Commit()
}

/*
Read retrieves notes associated with a specific task.
 1. Takes a database connection, and a parentID (parentID == area or task ID).
 2. Returns a slice of notes and an error if any occurs.
*/
func (tn *Note) Read(db *sql.DB, id int) ([]Note, error) {
	var notes []Note

	notesQuery := fmt.Sprintf(`
        SELECT notes.id, notes.title, notes.path, notes.type
        FROM notes
        JOIN bridge_notes ON notes.id = bridge_notes.note_id
        WHERE bridge_notes.note_id = ?
    `)

	rows, err := db.Query(notesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Type); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

// ReadByID retrieves a singular note by its ID
func (n *Note) ReadByID(db *sql.DB, noteID int) error {
	query := `
SELECT id, title, path, type
FROM notesTy
WHERE id = ?
`
	row := db.QueryRow(query, noteID)

	if err := row.Scan(&n.ID, &n.Title, &n.Path, &n.Type); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

/*
ReadAll retrieves all notes of a given type (areas or tasks)
 1. Takes a database connection, and a table name as parameters.
 2. Returns a slice of notes and an error if any occurs.
*/
func (tn *Note) ReadAll(db *sql.DB, noteType NoteType) ([]NoteTable, error) {
	var notesQuery string

	switch noteType {
	case TaskNoteType:
		notesQuery = `
	SELECT notes.id, notes.title, notes.path, tasks.title
	FROM notes
	INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
	INNER JOIN tasks ON tasks.ID = bridge_notes.parentID
	WHERE bridge_notes.cat_type = 1
	`
	case AreaNoteType:
		notesQuery = `
	SELECT notes.id, notes.title, notes.path, areas.title
	FROM notes
	INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
	INNER JOIN areas ON areas.ID = bridge_notes.parentID
	wHERE bridge_notes.cat_type = 2
	`
	default:
		return nil, fmt.Errorf("invalid note type: %d", noteType)
	}

	rows, err := db.Query(notesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []NoteTable
	for rows.Next() {
		var note NoteTable
		if err := rows.Scan(&note.NoteID, &note.NoteTitle, &note.NotePath, &note.LinkTitle); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
