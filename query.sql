-- name: CreateNote :execlastid
INSERT INTO notes (title, path) VALUES (?, ?)
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
    datetime(tasks.created_at) AS created_at,
    datetime(tasks.last_mod) AS last_mod,
    ROUND((julianday('now') - julianday(tasks.created_at)), 2) AS age_in_days,
    tasks.due_date,
    IFNULL(
        (SELECT GROUP_CONCAT(title, ', ') 
         FROM (SELECT DISTINCT notes.title 
               FROM notes 
               JOIN bridge_notes ON notes.id = bridge_notes.note_id 
               WHERE bridge_notes.parent_task_id = tasks.id)
        ), 
        ''
    ) AS note_title,
    programming_projects.path AS prog_proj,
    areas.title AS parent_area
FROM
    tasks
LEFT OUTER JOIN
    bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT OUTER JOIN
    notes ON notes.id = bridge_notes.note_id
LEFT OUTER JOIN
    prog_project_links ON tasks.id = prog_project_links.parent_task_id
LEFT OUTER JOIN
    programming_projects ON prog_project_links.project_id = programming_projects.id
LEFT OUTER JOIN
    areas ON tasks.area_id = areas.id
WHERE
    tasks.id = ?;


-- name: ReadTasks :many
SELECT 
    tasks.id, 
    tasks.title, 
    tasks.priority, 
    tasks.status, 
    tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)), 2) AS age_in_days,
    IFNULL(
        (SELECT GROUP_CONCAT(title, ', ') 
         FROM (SELECT DISTINCT notes.title 
               FROM notes 
               JOIN bridge_notes ON notes.id = bridge_notes.note_id 
               WHERE bridge_notes.parent_task_id = tasks.id)
        ), 
        ''
    ) AS note_titles,
    pp.path, 
    area.title AS parent_area
FROM 
    tasks
LEFT OUTER JOIN 
    bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT OUTER JOIN 
    notes ON bridge_notes.note_id = notes.id
LEFT OUTER JOIN 
    prog_project_links pjl ON pjl.parent_task_id = tasks.id
LEFT OUTER JOIN 
    programming_projects pp ON pjl.project_id = pp.id
LEFT OUTER JOIN 
    areas area ON area.id = tasks.area_id
GROUP BY 
    tasks.id;

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


-- name: ReadNoteByIDs :many
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE notes.id in (sqlc.slice(ids));

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


-- name: ReadAllNotes :many
SELECT notes.id, notes.title, notes.path, coalesce(tasks.title, areas.title) [area_or_task_title], case when bridge_notes.parent_cat = 1 then 'Task' else 'Area' end as [parent_type]
FROM notes
INNER JOIN bridge_notes ON bridge_notes.note_id = notes.id
LEFT JOIN tasks ON tasks.ID = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
LEFT JOIN areas ON areas.ID = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2;

-- name: UpdateAreaStatus :execresult
UPDATE areas SET status = ?  where id = ?
returning *;

-- name: UpdateAreaArchived :execresult
UPDATE areas SET archived = ?  where id = ?
returning *;

-- name: UpdateAreaTitle :execlastid
UPDATE areas set title = ? where id = ?
returning id;


-- name: UpdateTaskStatus :execresult
UPDATE tasks SET status = ?  where id = ?
returning *;

-- name: UpdateTaskPriority :execresult
UPDATE tasks SET priority = ?  where id = ?
returning *;

-- name: UpdateTaskTitle :execresult
UPDATE tasks set title = ? where id = ?
returning *;

-- name: UpdateTaskArchived :execresult
UPDATE tasks SET archived = ? WHERE id = ?
returning *;

-- name: UpdateTaskArea :execresult
UPDATE tasks set area_id = ? where id = ?
returning *;


-- name: DeleteNote :one
DELETE FROM notes WHERE id = ?
returning *;

-- name: DeleteNotes :execresult
DELETE FROM notes WHERE id in (sqlc.slice(ids))
returning *;

-- name: CreateArea :execlastid
INSERT INTO areas (title, status, archived)
VALUES (?, ?, ?)
returning id;

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


-- name: DeleteNotesFromSingleArea :execresult
DELETE FROM notes
WHERE notes.id IN (
		SELECT bridge_notes.note_id
		FROM bridge_notes
		WHERE parent_cat = 2 AND parent_area_id = ?
);

-- name: DeleteNotesFromMultipleAreas :execresult
DELETE FROM notes
WHERE id IN (
    SELECT note_id
    FROM bridge_notes
    WHERE parent_cat = 2 AND parent_area_id IN (sqlc.slice(ids))
)
RETURNING *;

-- name: DeleteSingleArea :one
DELETE FROM areas WHERE id = ?
returning id
;


-- name: DeleteMultipleAreas :execresult
DELETE FROM areas WHERE id IN (sqlc.slice(ids))
returning *;


-- name: CreateTask :execlastid
INSERT INTO tasks (
    title, priority, status, archived, due_date, area_id,
    created_at, last_mod
)
VALUES (
    ?, ?, ?, ?, ?, ?,
    datetime(current_timestamp, 'localtime'),
    datetime(current_timestamp, 'localtime')
)
returning id;

-- name: ReadTaskNote :many
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id = ? 
AND bridge_notes.parent_cat = 1;


-- name: ReadTaskNotes :execrows
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id in (sqlc.slice(ids))
AND bridge_notes.parent_cat = 1;

-- name: ReadAllTasks :many
	SELECT tasks.id, tasks.title, tasks.priority, tasks.status, tasks.archived,
    ROUND((julianday('now') - julianday(tasks.created_at)),2) AS age_in_days,
		IFNULL(GROUP_CONCAT(notes.title, ', '), '') as note_titles
	FROM tasks
	LEFT OUTER JOIN bridge_notes ON tasks.id = bridge_notes.parent_task_id AND bridge_notes.parent_cat = 1
	LEFT OUTER JOIN notes ON bridge_notes.note_id = notes.id
	GROUP BY tasks.id;


-- name: DeleteTask :one
DELETE FROM tasks
WHERE id = ?
returning id;


-- name: DeleteTasks :execrows
DELETE FROM tasks
WHERE id in (sqlc.slice(ids))
;


-- name: ReadAreaNote :many
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_task_id = ?
AND bridge_notes.parent_cat = 2;

-- name: ReadAreaNotes :execrows
SELECT notes.id, notes.title, notes.path, bridge_notes.parent_cat as type
FROM notes
INNER JOIN bridge_notes on notes.id = bridge_notes.note_id
WHERE bridge_notes.parent_area_id in (sqlc.slice(ids))
AND bridge_notes.parent_cat = 2
;

-- name: ReadAreas :many
SELECT 
    areas.id, areas.title, areas.status, areas.archived,
    IFNULL(GROUP_CONCAT(notes.title, ', '), '') AS note_titles, pp.path
FROM 
    areas
LEFT JOIN 
    bridge_notes ON areas.id = bridge_notes.parent_area_id AND bridge_notes.parent_cat = 2
LEFT JOIN 
    notes ON bridge_notes.note_id = notes.id
LEFT OUTER JOIN prog_project_links pjl ON pjl.parent_area_id = areas.id
LEFT OUTER JOIN programming_projects pp ON pjl.project_id = pp.id
GROUP BY 
    areas.id;

-- name: ReadAllProgProjects :many
SELECT path
FROM programming_projects;

-- name: CheckProgProjectExists :one
SELECT
  CASE WHEN EXISTS (
    SELECT 1
    FROM programming_projects
    WHERE path = ?
  ) THEN 1 ELSE 0 END AS prog_proj_exists;


-- name: FindProgProjectsForTask :one
SELECT pp.*
FROM programming_projects pp
JOIN prog_project_links pl on pp.id = pl.project_id
WHERE pl.parent_task_id = ?;

-- name: FindProgProjectsForArea :many
SELECT pp.*
FROM programming_projects pp
JOIN prog_project_links pl on pp.id = pl.project_id
WHERE pl.parent_area_id = ?;

-- name: InsertProgProject :one
INSERT INTO programming_projects (path)
VALUES (?)
RETURNING id;

-- name: CreateProjectTaskLink :exec
INSERT INTO prog_project_links (project_id, parent_cat, parent_task_id)
VALUES (?, ?, ?)
;

-- name: CreateProjectAreaLink :exec
INSERT INTO prog_project_links (project_id, parent_cat, parent_area_id)
VALUES (?, ?, ?)
;
