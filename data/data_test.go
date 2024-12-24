package data_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var dbConn *sql.DB

// TestMain is the entry point for testing. It allows setup and teardown before and after tests.
func TestMain(m *testing.M) {
	var err error
	dbConn, err = sql.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to open database: " + err.Error())
	}

	// Create tables and insert test data
	_, err = dbConn.Exec(`
PRAGMA foreign_keys = ON;
CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT, priority TEXT, status TEXT, archived BOOLEAN, created_at DATETIME, last_mod DATETIME, due_date DATETIME);
CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT, path TEXT);
CREATE TABLE areas (id INTEGER PRIMARY KEY, title TEXT, status TEXT, archived BOOLEAN, created_at DATETIME, last_mod DATETIME, due_date DATETIME);
CREATE TABLE bridge_notes (
  note_id INTEGER,
  parent_cat INTEGER,
  parent_task_id INTEGER,
  parent_area_id INTEGER,
  FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CHECK (parent_cat IN (1, 2)),
  FOREIGN KEY(parent_task_id) REFERENCES tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY(parent_area_id) REFERENCES areas(id) ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO tasks (id, title, priority, status, archived, created_at, last_mod, due_date) VALUES 
  (1, 'Test Task', 'High', 'Open', 0, '2023-01-01', '2023-01-01', '2023-01-01'),
  (2, 'Test Task 2', 'Medium', 'Open', 0, '2023-01-01', '2023-01-01', '2023-01-01');
INSERT INTO areas (id, title, status, archived, created_at, last_mod, due_date) VALUES 
  (1, 'Test Area', 'todo', 0, '2023-01-01', '2023-01-01', '2023-01-01'),
  (2, 'Test Area 2', 'todo', 0, '2023-01-01', '2023-01-01', '2023-01-01');
INSERT INTO notes (id, title, path) VALUES 
  (1, 'Note 1', '/path/to/note1'), 
  (2, 'Note 2', '/path/to/note2'),
  (3, 'Note 3', '/path/to/note3'), 
  (4, 'Note 4', '/path/to/note4');
INSERT INTO bridge_notes (note_id, parent_cat, parent_task_id) VALUES 
  (1, 1, 1),
  (2, 1, 1);
INSERT INTO bridge_notes (note_id, parent_cat, parent_area_id) VALUES 
  (3, 2, 1),
  (4, 2, 1);
`)`)
	if err != nil {
		panic("failed to insert test data: " + err.Error())
	}

	// Run the tests
	code := m.Run()

	// Teardown: Close the database connection and remove the test database file
	dbConn.Close()
	os.Remove("test.db")

	// Exit with the appropriate code
	os.Exit(code)
}
