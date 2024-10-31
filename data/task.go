package data

import (
	"database/sql"
	"time"
)

type CRUD interface {
	Create(db *sql.DB) error
	Read(db *sql.DB) (interface{}, error)
	ReadAll(db *sql.DB) ([]interface{}, error)
	Update(db *sql.DB) error
	Delete(db *sql.DB) error
	DeleteMultiple(db *sql.DB, ID ...int) error
}

type Task struct {
	ID             int
	Title          string
	Description    string
	Priority       string
	Status         string
	Archived       bool
	UpdateArchived bool
	CreatedAt      time.Time
	LastModified   time.Time
	DueDate        time.Time
	Notes          []Note
}

type Area struct {
	ID       int
	Title    string
	Type     string
	Deadline time.Time
	Tasks    []Task
	Status   string
	Archived bool
	Notes    []Note
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
