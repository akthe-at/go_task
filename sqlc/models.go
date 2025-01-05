// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"database/sql"
	"time"
)

type Area struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	Status    sql.NullString `json:"status"`
	Archived  bool           `json:"archived"`
	CreatedAt time.Time      `json:"created_at"`
	LastMod   time.Time      `json:"last_mod"`
}

type AreaID struct {
	ID int64 `json:"id"`
}

type BridgeNote struct {
	NoteID       sql.NullInt64 `json:"note_id"`
	ParentCat    sql.NullInt64 `json:"parent_cat"`
	ParentTaskID sql.NullInt64 `json:"parent_task_id"`
	ParentAreaID sql.NullInt64 `json:"parent_area_id"`
}

type Note struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type NoteID struct {
	ID int64 `json:"id"`
}

type ProgProjID struct {
	ID int64 `json:"id"`
}

type ProgProjectLink struct {
	ProjectID    sql.NullInt64 `json:"project_id"`
	ParentCat    sql.NullInt64 `json:"parent_cat"`
	ParentTaskID sql.NullInt64 `json:"parent_task_id"`
	ParentAreaID sql.NullInt64 `json:"parent_area_id"`
}

type ProgrammingProject struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

type Task struct {
	ID        int64          `json:"id"`
	Title     string         `json:"title"`
	Priority  sql.NullString `json:"priority"`
	Status    sql.NullString `json:"status"`
	Archived  bool           `json:"archived"`
	CreatedAt time.Time      `json:"created_at"`
	LastMod   time.Time      `json:"last_mod"`
	DueDate   sql.NullTime   `json:"due_date"`
	AreaID    sql.NullInt64  `json:"area_id"`
}

type TaskID struct {
	ID int64 `json:"id"`
}
