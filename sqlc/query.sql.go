// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package sqlc

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

const createArea = `-- name: CreateArea :execlastid
INSERT INTO areas (title, status, archived)
VALUES (?, ?, ?)
returning id, title, status, archived, created_at, last_mod
`

type CreateAreaParams struct {
	Title    string         `json:"title"`
	Status   sql.NullString `json:"status"`
	Archived bool           `json:"archived"`
}

func (q *Queries) CreateArea(ctx context.Context, arg CreateAreaParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createArea, arg.Title, arg.Status, arg.Archived)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createAreaBridgeNote = `-- name: CreateAreaBridgeNote :execlastid
INSERT INTO bridge_notes (note_id, parent_cat, parent_area_id) VALUES (?, ?, ?)
`

type CreateAreaBridgeNoteParams struct {
	NoteID       sql.NullInt64 `json:"note_id"`
	ParentCat    sql.NullInt64 `json:"parent_cat"`
	ParentAreaID sql.NullInt64 `json:"parent_area_id"`
}

func (q *Queries) CreateAreaBridgeNote(ctx context.Context, arg CreateAreaBridgeNoteParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createAreaBridgeNote, arg.NoteID, arg.ParentCat, arg.ParentAreaID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createNote = `-- name: CreateNote :execlastid
INSERT INTO notes (title, path) VALUES  (?, ?)
returning id
`

type CreateNoteParams struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

func (q *Queries) CreateNote(ctx context.Context, arg CreateNoteParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createNote, arg.Title, arg.Path)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createTask = `-- name: CreateTask :execlastid
INSERT INTO tasks (
    title, priority, status, archived, due_date,
    created_at, last_mod
)
VALUES (
    ?, ?, ?, ?, ?,
    datetime(current_timestamp, 'localtime'),
    datetime(current_timestamp, 'localtime')
)
returning id
`

type CreateTaskParams struct {
	Title    string         `json:"title"`
	Priority sql.NullString `json:"priority"`
	Status   sql.NullString `json:"status"`
	Archived bool           `json:"archived"`
	DueDate  sql.NullTime   `json:"due_date"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createTask,
		arg.Title,
		arg.Priority,
		arg.Status,
		arg.Archived,
		arg.DueDate,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const createTaskBridgeNote = `-- name: CreateTaskBridgeNote :execlastid
INSERT INTO bridge_notes (note_id, parent_cat, parent_task_id) VALUES (?, ?, ?)
`

type CreateTaskBridgeNoteParams struct {
	NoteID       sql.NullInt64 `json:"note_id"`
	ParentCat    sql.NullInt64 `json:"parent_cat"`
	ParentTaskID sql.NullInt64 `json:"parent_task_id"`
}

func (q *Queries) CreateTaskBridgeNote(ctx context.Context, arg CreateTaskBridgeNoteParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createTaskBridgeNote, arg.NoteID, arg.ParentCat, arg.ParentTaskID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const deleteArea = `-- name: DeleteArea :execresult
DELETE FROM areas WHERE id = ?
`

func (q *Queries) DeleteArea(ctx context.Context, id int64) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteArea, id)
}

const deleteAreaAndNotes = `-- name: DeleteAreaAndNotes :execresult
DELETE FROM notes
WHERE notes.id IN (
		SELECT bridge_notes.note_id
		FROM bridge_notes
		WHERE parent_cat = 2 AND parent_area_id = ?
)
`

func (q *Queries) DeleteAreaAndNotes(ctx context.Context, parentAreaID sql.NullInt64) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteAreaAndNotes, parentAreaID)
}

const deleteAreasAndNotesMultiple = `-- name: DeleteAreasAndNotesMultiple :execresult
DELETE FROM notes
WHERE id IN (
    SELECT note_id
    FROM bridge_notes
    WHERE parent_cat = 2 AND parent_area_id IN (?)
)
`

func (q *Queries) DeleteAreasAndNotesMultiple(ctx context.Context, parentAreaID sql.NullInt64) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteAreasAndNotesMultiple, parentAreaID)
}

const deleteMultipleAreas = `-- name: DeleteMultipleAreas :execresult
;

DELETE FROM areas WHERE id IN (?)
returning id, title, status, archived, created_at, last_mod
`

func (q *Queries) DeleteMultipleAreas(ctx context.Context, id int64) (sql.Result, error) {
	return q.db.ExecContext(ctx, deleteMultipleAreas, id)
}

const deleteNote = `-- name: DeleteNote :one
DELETE FROM notes WHERE id = ?
returning id, title, path
`

func (q *Queries) DeleteNote(ctx context.Context, id int64) (Note, error) {
	row := q.db.QueryRowContext(ctx, deleteNote, id)
	var i Note
	err := row.Scan(&i.ID, &i.Title, &i.Path)
	return i, err
}

const deleteNotes = `-- name: DeleteNotes :execresult
DELETE FROM notes WHERE id in (/*SLICE:ids*/?)
returning id, title, path
`

func (q *Queries) DeleteNotes(ctx context.Context, ids []int64) (sql.Result, error) {
	query := deleteNotes
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	return q.db.ExecContext(ctx, query, queryParams...)
}

const deleteTask = `-- name: DeleteTask :one
DELETE FROM tasks
WHERE id = ?
returning id
`

func (q *Queries) DeleteTask(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, deleteTask, id)
	err := row.Scan(&id)
	return id, err
}

const deleteTasks = `-- name: DeleteTasks :execrows
DELETE FROM tasks
WHERE id in (/*SLICE:ids*/?)
`

func (q *Queries) DeleteTasks(ctx context.Context, ids []int64) (int64, error) {
	query := deleteTasks
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	result, err := q.db.ExecContext(ctx, query, queryParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const readAllAreaNotes = `-- name: ReadAllAreaNotes :many
SELECT notes.id, notes.title, notes.path, areas.title as area_title, areas.id as parent_id
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
INNER JOIN areas ON areas.ID = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2
`

type ReadAllAreaNotesRow struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	AreaTitle string `json:"area_title"`
	ParentID  int64  `json:"parent_id"`
}

func (q *Queries) ReadAllAreaNotes(ctx context.Context) ([]ReadAllAreaNotesRow, error) {
	rows, err := q.db.QueryContext(ctx, readAllAreaNotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAllAreaNotesRow
	for rows.Next() {
		var i ReadAllAreaNotesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Path,
			&i.AreaTitle,
			&i.ParentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readAllNotes = `-- name: ReadAllNotes :many
SELECT notes.id, notes.title, notes.path, coalesce(tasks.title, areas.title) [area_or_task_title], case when bridge_notes.parent_cat = 1 then 'Task' else 'Area' end as [parent_type]
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
LEFT JOIN tasks ON tasks.ID = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT JOIN areas ON areas.ID = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2
`

type ReadAllNotesRow struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Path            string `json:"path"`
	AreaOrTaskTitle string `json:"[area_or_task_title]"`
	ParentType      string `json:"[parent_type]"`
}

func (q *Queries) ReadAllNotes(ctx context.Context) ([]ReadAllNotesRow, error) {
	rows, err := q.db.QueryContext(ctx, readAllNotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAllNotesRow
	for rows.Next() {
		var i ReadAllNotesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Path,
			&i.AreaOrTaskTitle,
			&i.ParentType,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readAllTaskNotes = `-- name: ReadAllTaskNotes :many
SELECT notes.id, notes.title, notes.path, tasks.title as task_title, tasks.id  as parent_id
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
INNER JOIN tasks ON tasks.ID = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
`

type ReadAllTaskNotesRow struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	TaskTitle string `json:"task_title"`
	ParentID  int64  `json:"parent_id"`
}

func (q *Queries) ReadAllTaskNotes(ctx context.Context) ([]ReadAllTaskNotesRow, error) {
	rows, err := q.db.QueryContext(ctx, readAllTaskNotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAllTaskNotesRow
	for rows.Next() {
		var i ReadAllTaskNotesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Path,
			&i.TaskTitle,
			&i.ParentID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readAllTasks = `-- name: ReadAllTasks :many
	SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)),2) AS age_in_days,
		IFNULL(GROUP_CONCAT(notes.title, ', '), '') as note_titles
	FROM tasks
	LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
	LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
	GROUP BY tasks.id
`

type ReadAllTasksRow struct {
	ID         int64          `json:"id"`
	Title      string         `json:"title"`
	Priority   sql.NullString `json:"priority"`
	Status     sql.NullString `json:"status"`
	Archived   bool           `json:"archived"`
	AgeInDays  float64        `json:"age_in_days"`
	NoteTitles interface{}    `json:"note_titles"`
}

func (q *Queries) ReadAllTasks(ctx context.Context) ([]ReadAllTasksRow, error) {
	rows, err := q.db.QueryContext(ctx, readAllTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAllTasksRow
	for rows.Next() {
		var i ReadAllTasksRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Priority,
			&i.Status,
			&i.Archived,
			&i.AgeInDays,
			&i.NoteTitles,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readArea = `-- name: ReadArea :one
SELECT 
    areas.id, areas.title, areas.status, areas.archived,
    notes.id, notes.title, notes.path
FROM 
    areas
LEFT JOIN 
    bridge_notes ON areas.id = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2
LEFT JOIN 
    notes ON bridge_notes.note_id = notes.id
WHERE 
    areas.id = ?
`

type ReadAreaRow struct {
	ID       int64          `json:"id"`
	Title    string         `json:"title"`
	Status   sql.NullString `json:"status"`
	Archived bool           `json:"archived"`
	ID_2     sql.NullInt64  `json:"id_2"`
	Title_2  sql.NullString `json:"title_2"`
	Path     sql.NullString `json:"path"`
}

func (q *Queries) ReadArea(ctx context.Context, id int64) (ReadAreaRow, error) {
	row := q.db.QueryRowContext(ctx, readArea, id)
	var i ReadAreaRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Status,
		&i.Archived,
		&i.ID_2,
		&i.Title_2,
		&i.Path,
	)
	return i, err
}

const readAreas = `-- name: ReadAreas :many
;

SELECT 
    areas.id, areas.title, areas.status, areas.archived,
    IFNULL(GROUP_CONCAT(notes.title, ', '), '') AS note_titles
FROM 
    areas
LEFT JOIN 
    bridge_notes ON areas.id = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2
LEFT JOIN 
    notes ON bridge_notes.note_id = notes.id
GROUP BY 
    areas.id
`

type ReadAreasRow struct {
	ID         int64          `json:"id"`
	Title      string         `json:"title"`
	Status     sql.NullString `json:"status"`
	Archived   bool           `json:"archived"`
	NoteTitles interface{}    `json:"note_titles"`
}

func (q *Queries) ReadAreas(ctx context.Context) ([]ReadAreasRow, error) {
	rows, err := q.db.QueryContext(ctx, readAreas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadAreasRow
	for rows.Next() {
		var i ReadAreasRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Status,
			&i.Archived,
			&i.NoteTitles,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readNote = `-- name: ReadNote :many
SELECT notes.id, notes.title, bridge_notes.parent_cat as type
FROM notes
JOIN bridge_notes ON notes.id = bridge_notes.note_id
WHERE bridge_notes.note_id = ?
`

type ReadNoteRow struct {
	ID    int64         `json:"id"`
	Title string        `json:"title"`
	Type  sql.NullInt64 `json:"type"`
}

func (q *Queries) ReadNote(ctx context.Context, noteID sql.NullInt64) ([]ReadNoteRow, error) {
	rows, err := q.db.QueryContext(ctx, readNote, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadNoteRow
	for rows.Next() {
		var i ReadNoteRow
		if err := rows.Scan(&i.ID, &i.Title, &i.Type); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readNoteByID = `-- name: ReadNoteByID :one
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE notes.id = ?
`

type ReadNoteByIDRow struct {
	ID    int64         `json:"id"`
	Title string        `json:"title"`
	Path  string        `json:"path"`
	Type  sql.NullInt64 `json:"type"`
}

func (q *Queries) ReadNoteByID(ctx context.Context, id int64) (ReadNoteByIDRow, error) {
	row := q.db.QueryRowContext(ctx, readNoteByID, id)
	var i ReadNoteByIDRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Path,
		&i.Type,
	)
	return i, err
}

const readTask = `-- name: ReadTask :one
SELECT
    tasks.id AS task_id,
    tasks.title AS task_title,
    tasks.priority,
    tasks.status,
    tasks.archived,
    tasks.created_at,
    tasks.last_mod,
    ROUND((julianday('now') - julianday(tasks.created_at)), 2) AS age_in_days,
    tasks.due_date,
		IFNULL(GROUP_CONCAT(notes.title, ', '), '') as note_title
FROM 
    tasks
LEFT JOIN 
    bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT JOIN 
    notes ON notes.id = bridge_notes.note_id
WHERE 
    tasks.id = ?
`

type ReadTaskRow struct {
	TaskID    int64          `json:"task_id"`
	TaskTitle string         `json:"task_title"`
	Priority  sql.NullString `json:"priority"`
	Status    sql.NullString `json:"status"`
	Archived  bool           `json:"archived"`
	CreatedAt time.Time      `json:"created_at"`
	LastMod   time.Time      `json:"last_mod"`
	AgeInDays float64        `json:"age_in_days"`
	DueDate   sql.NullTime   `json:"due_date"`
	NoteTitle interface{}    `json:"note_title"`
}

func (q *Queries) ReadTask(ctx context.Context, id int64) (ReadTaskRow, error) {
	row := q.db.QueryRowContext(ctx, readTask, id)
	var i ReadTaskRow
	err := row.Scan(
		&i.TaskID,
		&i.TaskTitle,
		&i.Priority,
		&i.Status,
		&i.Archived,
		&i.CreatedAt,
		&i.LastMod,
		&i.AgeInDays,
		&i.DueDate,
		&i.NoteTitle,
	)
	return i, err
}

const readTaskNote = `-- name: ReadTaskNote :many
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id = ? 
AND bridge_notes.parent_cat = 1
`

type ReadTaskNoteRow struct {
	ID    int64         `json:"id"`
	Title string        `json:"title"`
	Path  string        `json:"path"`
	Type  sql.NullInt64 `json:"type"`
}

func (q *Queries) ReadTaskNote(ctx context.Context, parentTaskID sql.NullInt64) ([]ReadTaskNoteRow, error) {
	rows, err := q.db.QueryContext(ctx, readTaskNote, parentTaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadTaskNoteRow
	for rows.Next() {
		var i ReadTaskNoteRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Path,
			&i.Type,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const readTasks = `-- name: ReadTasks :many
SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)), 2) AS age_in_days,
    IFNULL(GROUP_CONCAT(notes.title, ', '), '') AS note_titles
FROM tasks
LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
GROUP BY tasks.id
`

type ReadTasksRow struct {
	ID         int64          `json:"id"`
	Title      string         `json:"title"`
	Priority   sql.NullString `json:"priority"`
	Status     sql.NullString `json:"status"`
	Archived   bool           `json:"archived"`
	AgeInDays  float64        `json:"age_in_days"`
	NoteTitles interface{}    `json:"note_titles"`
}

func (q *Queries) ReadTasks(ctx context.Context) ([]ReadTasksRow, error) {
	rows, err := q.db.QueryContext(ctx, readTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadTasksRow
	for rows.Next() {
		var i ReadTasksRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Priority,
			&i.Status,
			&i.Archived,
			&i.AgeInDays,
			&i.NoteTitles,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
