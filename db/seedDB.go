package db

func SeedDB() string {
	db_seed := `INSERT INTO tasks (id, title, priority, status, archived, created_at, last_mod, due_date) VALUES
(1, 'Task 1', 'High', 1, 0, datetime('2023-10-01 10:00:00'), datetime('2023-10-01 10:00:00'), datetime('2023-10-10 10:00:00')),
(2, 'Task 2', 'Medium', 0, 0, datetime('2023-10-02 11:00:00'), datetime('2023-10-02 11:00:00'), datetime('2023-10-11 11:00:00')),
(3, 'Task 3', 'Low', 1, 1, datetime('2023-10-03 12:00:00'), datetime('2023-10-03 12:00:00'), datetime('2023-10-12 12:00:00')),
(4, 'Task 4', 'High', 0, 0, datetime('2023-10-04 13:00:00'), datetime('2023-10-04 13:00:00'), datetime('2023-10-13 13:00:00')),
(5, 'Task 5', 'Medium', 1, 0, datetime('2023-10-05 14:00:00'), datetime('2023-10-05 14:00:00'), datetime('2023-10-14 14:00:00')),
(6, 'Task 6', 'Low', 0, 1, datetime('2023-10-06 15:00:00'), datetime('2023-10-06 15:00:00'), datetime('2023-10-15 15:00:00')),
(7, 'Task 7', 'High', 1, 0, datetime('2023-10-07 16:00:00'), datetime('2023-10-07 16:00:00'), datetime('2023-10-16 16:00:00')),
(8, 'Task 8', 'Medium', 0, 0, datetime('2023-10-08 17:00:00'), datetime('2023-10-08 17:00:00'), datetime('2023-10-17 17:00:00')),
(9, 'Task 9', 'Low', 1, 1, datetime('2023-10-09 18:00:00'), datetime('2023-10-09 18:00:00'), datetime('2023-10-18 18:00:00')),
(10, 'Task 10', 'High', 0, 0, datetime('2023-10-10 19:00:00'), datetime('2023-10-10 19:00:00'), datetime('2023-10-19 19:00:00'));
`
	return db_seed
}
