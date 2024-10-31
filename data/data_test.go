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
	CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT, description TEXT, priority TEXT, status TEXT, archived BOOLEAN, created_at DATETIME, last_mod DATETIME, due_date DATETIME);
	CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT, path TEXT);
	CREATE TABLE task_notes (note_id INTEGER, task_id INTEGER);
	CREATE TABLE areas (id INTEGER PRIMARY KEY, title TEXT, type TEXT, deadline DATETIME, status TEXT, archived BOOLEAN);
	CREATE TABLE area_notes (note_id INTEGER, area_id INTEGER);
	INSERT INTO tasks (id, title, description, priority, status, archived, created_at, last_mod, due_date) VALUES (1, 'Test Task', 'Description', 'High', 'Open', 0, '2023-01-01', '2023-01-01', '2023-01-01');
INSERT INTO tasks (id, title, description, priority, status, archived, created_at, last_mod, due_date) VALUES (2, 'Test Task 2', 'Description 2', 'Medium', 'Open', 0, '2023-01-01', '2023-01-01', '2023-01-01');
	INSERT INTO notes (id, title, path) VALUES (1, 'Note 1', '/path/to/note1'), (2, 'Note 2', '/path/to/note2');
INSERT INTO notes (id, title, path) VALUES (3, 'Note 3', '/path/to/note3'), (4, 'Note 4', '/path/to/note4');
	INSERT INTO task_notes (note_id, task_id) VALUES (1, 1), (2, 1);
INSERT INTO task_notes (note_id, task_id) VALUES (3, 2), (4, 2);
	INSERT INTO areas (id, title, type, deadline, status, archived) VALUES (1, 'Test Area', 'Type', '2023-01-01', 'Open', 0);
INSERT INTO areas (id, title, type, deadline, status, archived) VALUES (2, 'Test Area 2', 'Type 2', '2023-01-01', 'Open', 0);
	INSERT INTO area_notes (note_id, area_id) VALUES (1, 1), (2, 1);
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

func TestTaskTable_Read(t *testing.T) {
	taskTable := &data.TaskTable{Task: data.Task{ID: 1}}
	task, err := taskTable.Read(dbConn)
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

func TestTaskTable_ReadAll(t *testing.T) {
	taskTable := &data.TaskTable{}
	tasks, err := taskTable.ReadAll(dbConn)
	if err != nil {
		t.Fatalf("failed to read all tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestAreaTable_Read(t *testing.T) {
	areaTable := &data.AreaTable{Area: data.Area{ID: 1}}
	area, err := areaTable.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read area: %v", err)
	}

	if area.ID != 1 {
		t.Errorf("expected area ID 1, got %d", area.ID)
	}
	if len(area.Notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(area.Notes))
	}
}

func TestAreaTable_ReadAll(t *testing.T) {
	areaTable := &data.AreaTable{}
	areas, err := areaTable.ReadAll(dbConn)
	if err != nil {
		t.Fatalf("failed to read all areas: %v", err)
	}

	if len(areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(areas))
	}
}

func TestTaskTable_Update(t *testing.T) {
	taskTable := &data.TaskTable{Task: data.Task{ID: 1, Title: "Updated Task", Description: "Updated Description"}}
	_, err := taskTable.Update(dbConn)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	task, err := taskTable.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}

	if task.Title != "Updated Task" {
		t.Errorf("expected task title 'Updated Task', got %s", task.Title)
	}
	if task.Description != "Updated Description" {
		t.Errorf("expected task description 'Updated Description', got %s", task.Description)
	}
}

func TestTaskTable_Delete(t *testing.T) {
	taskTable := &data.TaskTable{Task: data.Task{ID: 1}}
	err := taskTable.Delete(dbConn)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	_, err = taskTable.Read(dbConn)
	if err == nil {
		t.Fatalf("expected error when reading deleted task, got nil")
	}
}

func TestAreaTable_Update(t *testing.T) {
	areaTable := &data.AreaTable{Area: data.Area{ID: 1, Title: "Updated Area", Type: "Updated Type"}}
	_, err := areaTable.Update(dbConn)
	if err != nil {
		t.Fatalf("failed to update area: %v", err)
	}

	area, err := areaTable.Read(dbConn)
	if err != nil {
		t.Fatalf("failed to read area: %v", err)
	}

	if area.Title != "Updated Area" {
		t.Errorf("expected area title 'Updated Area', got %s", area.Title)
	}
	if area.Type != "Updated Type" {
		t.Errorf("expected area type 'Updated Type', got %s", area.Type)
	}
}

func TestAreaTable_Delete(t *testing.T) {
	areaTable := &data.AreaTable{Area: data.Area{ID: 1}}
	err := areaTable.Delete(dbConn)
	if err != nil {
		t.Fatalf("failed to delete area: %v", err)
	}

	_, err = areaTable.Read(dbConn)
	if err == nil {
		t.Fatalf("expected error when reading deleted area, got nil")
	}
}

func TestTaskTable_DeleteWithNullValues(t *testing.T) {
	taskTable := &data.TaskTable{Task: data.Task{ID: 0}}
	err := taskTable.Delete(dbConn)
	if err == nil {
		t.Fatalf("expected error when deleting task with null ID, got nil")
	}
}

func TestAreaTable_DeleteWithNullValues(t *testing.T) {
	areaTable := &data.AreaTable{Area: data.Area{ID: 0}}
	err := areaTable.Delete(dbConn)
	if err == nil {
		t.Fatalf("expected error when deleting area with null ID, got nil")
	}
}
