PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS areas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    status TEXT,
    archived BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
    last_mod TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime'))
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    priority TEXT,
    status TEXT,
    archived BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
    last_mod TEXT NOT NULL DEFAULT (datetime(current_timestamp, 'localtime')),
    due_date TEXT,
    area_id INTEGER,
    FOREIGN KEY(area_id) REFERENCES areas(id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE TRIGGER update_last_mod_tasks
BEFORE UPDATE ON tasks
FOR EACH ROW
BEGIN
    UPDATE tasks SET last_mod = (datetime(current_timestamp, 'localtime')) WHERE id = OLD.id;
END;

CREATE TRIGGER update_last_mod_areas
BEFORE UPDATE ON areas
FOR EACH ROW
BEGIN
    UPDATE areas SET last_mod = (datetime(current_timestamp, 'localtime')) WHERE id = OLD.id;
END;

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS bridge_notes (
    note_id INTEGER,
    parent_cat INTEGER,
    parent_task_id INTEGER,
    parent_area_id INTEGER,
    FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CHECK (parent_cat IN (1, 2)),
    FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS programming_projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS prog_project_links (
    project_id INTEGER,
    parent_cat INTEGER,
    parent_task_id INTEGER, 
    parent_area_id INTEGER,
    FOREIGN KEY(project_id) REFERENCES programming_projects(id) ON DELETE CASCADE,
    CHECK (parent_cat IN (1, 2)),
    FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE
);

