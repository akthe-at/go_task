PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS areas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    status TEXT,
    archived BOOLEAN
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    priority TEXT,
    status TEXT,
    archived BOOLEAN,
    created_at DATETIME,
    last_mod DATETIME,
    due_date DATETIME,
    area_id INTEGER,
    FOREIGN KEY(area_id) REFERENCES areas(id) ON DELETE SET NULL ON UPDATE CASCADE
);

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
