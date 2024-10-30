package data

import "time"

type Task struct {
	ID           int
	Title        string
	Description  string
	Priority     string
	Status       string
	Archived     bool
	CreatedAt    time.Time
	LastModified time.Time
	DueDate      time.Time
}

type Area struct {
	ID       int
	Title    string
	Type     string
	Deadline time.Time
	Tasks    []Task
	Status   string
	Archived bool
}
