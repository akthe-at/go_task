-- name: CreateNote :execlastid
INSERT INTO notes (title, path) VALUES  (?, ?)
returning id;

-- name: CreateTaskBridgeNote :execlastid
INSERT INTO bridge_notes (note_id, parent_cat, parent_task_id) VALUES (?, ?, ?);

-- name: CreateAreaBridgeNote :execlastid
INSERT INTO bridge_notes (note_id, parent_cat, parent_area_id) VALUES (?, ?, ?);

-- name: ReadTask :one
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
    tasks.id = ?;

-- name: ReadTasks :many
SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)), 2) AS age_in_days,
    IFNULL(GROUP_CONCAT(notes.title, ', '), '') AS note_titles
FROM tasks
LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
GROUP BY tasks.id;

-- name: ReadNote :many
SELECT notes.id, notes.title, bridge_notes.parent_cat as type
FROM notes
JOIN bridge_notes ON notes.id = bridge_notes.note_id
WHERE bridge_notes.note_id = ?;

-- name: ReadNoteByID :one
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE notes.id = ?;

-- name: ReadAllTaskNotes :many
SELECT notes.id, notes.title, notes.path, tasks.title as task_title, tasks.id  as parent_id
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
INNER JOIN tasks ON tasks.ID = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1;


-- name: ReadAllAreaNotes :many
SELECT notes.id, notes.title, notes.path, areas.title as area_title, areas.id as parent_id
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
INNER JOIN areas ON areas.ID = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2;


-- name: DeleteNote :one
DELETE FROM notes WHERE id = ?
returning *;

-- name: DeleteNotes :many
DELETE FROM notes WHERE id IN (?)
returning *;

-- name: CreateArea :execlastid
INSERT INTO areas (title, status, archived)
VALUES (?, ?, ?)
returning *;

-- name: ReadArea :one
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
    areas.id = ?;


-- name: DeleteAreaAndNotes :execresult
DELETE FROM notes
WHERE notes.id IN (
		SELECT bridge_notes.note_id
		FROM bridge_notes
		WHERE parent_cat = 2 AND parent_area_id = ?
);
DELETE FROM areas WHERE id = ?;

-- name: DeleteAreasAndNotesMultiple :execresult
DELETE FROM notes
WHERE id IN (
    SELECT note_id
    FROM bridge_notes
    WHERE parent_cat = 2 AND parent_area_id IN (?)
)
;
DELETE FROM areas 
WHERE id IN (?)
RETURNING *;

-- name: DeleteArea :execresult
DELETE FROM areas WHERE id = ?
;

-- name: DeleteMultipleAreas :execresult
DELETE FROM areas WHERE id IN (?)
returning *;

-- name: CreateTask :execlastid
INSERT INTO tasks (title, priority, status, archived, created_at, last_mod, due_date)
VALUES (?, ?, ?, ?, ?, ?, ?)
returning *;

-- name: ReadTaskNote :many
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id = ? 
AND bridge_notes.parent_cat = 1;

-- name: ReadAllTasks :many
	SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)),2) AS age_in_days,
		IFNULL(GROUP_CONCAT(notes.title, ', '), '') as note_titles
	FROM tasks
	LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
	LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
	GROUP BY tasks.id;


-- name: DeleteTask :execresult
DELETE FROM notes
WHERE id IN (
    SELECT note_id
    FROM bridge_notes
    WHERE parent_task_id = ? AND parent_cat = 1
);
DELETE FROM tasks
WHERE id = ?
AND NOT EXISTS (
    SELECT 1
    FROM bridge_notes
    WHERE parent_task_id = ? AND parent_cat = 1
);

-- name: DeleteTaskAndNotes :execresult
DELETE FROM notes
WHERE id IN (
    SELECT note_id
    FROM bridge_notes
    WHERE parent_cat = 1 AND parent_task_id = ?
)
RETURNING *;

DELETE FROM tasks
WHERE id = ?
RETURNING *;

-- name: ReadAreas :many
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
    areas.id;
