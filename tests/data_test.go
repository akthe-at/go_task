package data_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/akthe-at/go_task/data"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
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
`)
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

func TestTask_Read(t *testing.T) {
	task := &data.Task{ID: 1}
	err := task.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	if task.ID != 1 {
		t.Errorf("expected task ID 1, got %d", task.ID)
	}
	if len(task.Notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(task.Notes))
	}
}

func TestTask_ReadAll(t *testing.T) {
	task := &data.Task{}
	tasks, err := task.ReadAll(dbConn)
	if err != nil {
		t.Fatalf("failed to read all tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestArea_Read(t *testing.T) {
	area := &data.Area{ID: 1}
	err := area.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read area: %v", err)
	}

	if area.ID != 1 {
		t.Errorf("expected area ID 1, got %d", area.ID)
	}
	if len(area.Notes) != 2 {
		t.Errorf("expected 2 notes, got %d, print: %v", len(area.Notes), area)
	}
}

func TestArea_ReadAll(t *testing.T) {
	area := &data.Area{}
	areas, err := area.ReadAll(dbConn)
	if err != nil {
		t.Fatalf("failed to read all areas: %v", err)
	}

	if len(areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(areas))
	}
}

func TestTask_Update(t *testing.T) {
	task := &data.Task{ID: 1, Title: "Updated Task"}
	results, err := task.Update(dbConn)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	testResults, err := results.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}

	if testResults != 1 {
		t.Fatalf("expected 1 row affected, got %d", testResults)
	}

	err = task.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	if task.Title != "Updated Task" {
		t.Errorf("expected task title 'Updated Task', got %s", task.Title)
	}
}

func TestTask_Delete(t *testing.T) {
	task := &data.Task{ID: 1}
	err := task.Delete(dbConn)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	err = task.Read(dbConn)
	if err == nil {
		t.Fatalf("expected error when reading deleted task, got nil")
	}
}

func TestArea_Update(t *testing.T) {
	area := &data.Area{ID: 1, Title: "Updated Area", Status: "Updated Status"}
	results, err := area.Update(dbConn)
	if err != nil {
		t.Fatalf("failed to update area: %v", err)
	}

	testResults, err := results.RowsAffected()
	if err != nil {
		t.Fatalf("failed to get rows affected: %v", err)
	}

	if testResults != 1 {
		t.Fatalf("expected 1 row affected, got %d", testResults)
	}

	err = area.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read area: %v", err)
	}

	if area.Title != "Updated Area" {
		t.Errorf("expected area title 'Updated Area', got %s", area.Title)
	}
	if area.Status != "Updated Status" {
		t.Errorf("expected area status 'Updated Status', got %s", area.Status)
	}
}

func TestArea_Delete(t *testing.T) {
	area := &data.Area{ID: 1}
	err := area.Delete(dbConn)
	if err != nil {
		t.Fatalf("failed to delete area: %v", err)
	}

	err = area.Read(dbConn)
	if err == nil {
		t.Fatalf("expected error when reading deleted area, got nil")
	}
}

func TestTask_DeleteWithNullValues(t *testing.T) {
	task := &data.Task{ID: 0}
	err := task.Delete(dbConn)
	if err == nil {
		t.Fatalf("expected error when deleting task with null ID, got nil")
	}
}

func TestArea_DeleteWithNullValues(t *testing.T) {
	area := &data.Area{ID: 0}
	err := area.Delete(dbConn)
	if err == nil {
		t.Fatalf("expected error when deleting area with null ID, got nil")
	}
}

func TestArea_DeleteMultiple(t *testing.T) {
	// Insert multiple areas for testing
	_, err := dbConn.Exec(`

INSERT INTO areas (id, title, status, archived, created_at, last_mod, due_date) VALUES 
	(3, 'Test Area 3','done',0, '2023-01-01', '2023-01-01', '2023-01-01'),
	(4, 'Test Area 4','done',0, '2023-01-01', '2023-01-01', '2023-01-01');
	`)
	if err != nil {
		t.Fatalf("failed to insert test areas: %v", err)
	}

	area := &data.Area{}
	err = area.DeleteMultiple(dbConn, []int{3, 4})
	if err != nil {
		t.Fatalf("failed to delete multiple areas: %v", err)
	}

	// Verify the areas are deleted
	for _, id := range []int{3, 4} {
		area := &data.Area{ID: id}
		err = area.Read(dbConn)
		if err == nil {
			t.Fatalf("expected error when reading deleted area with ID %d, got nil", id)
		}
	}
}
