package storage

import (
	"Todolistick/models"
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dataSourceName string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}
	Storage, createTableQuery := &SQLiteStorage{db: db}, `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        completed BOOLEAN
    );` // Создать таблицу, если её нет
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return Storage, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) GetAll() ([]models.Todo, error) {
	rows, err := s.db.Query("SELECT id, title, completed FROM Todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		Todos = append(Todos, todo)
	}

	return Todos, nil
}

func (s *SQLiteStorage) GetByID(id int) (models.Todo, error) {
	var Todo models.Todo
	err := s.db.QueryRow("SELECT id, title, completed FROM todos WHERE id = ?", id).Scan(&Todo.ID, &Todo.Title, &Todo.Completed)
	if err != nil {
		if err == sql.ErrNoRows {
			return Todo, errors.New("Todo not found")
		}
		return Todo, err

	}
	return Todo, nil
}

func (s *SQLiteStorage) Add(todo models.Todo) (models.Todo, error) {
	result, err := s.db.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", todo.Title, todo.Completed)
	if err != nil {
		return todo, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return todo, err
	}
	todo.ID = int(id)
	return todo, nil
}

func (s *SQLiteStorage) Update(todo models.Todo) error {
	_, err := s.db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", todo.Title, todo.Completed, todo.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}
