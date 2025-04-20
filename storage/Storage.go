package storage

import (
	"Todolistick/models"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			title TEXT,
			completed BOOLEAN
		);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) GetAll() ([]models.Todo, error) {
	rows, err := s.db.Query("SELECT id, title, completed FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (s *PostgresStorage) GetByID(id int) (models.Todo, error) {
	var todo models.Todo
	err := s.db.QueryRow("SELECT id, title, completed FROM todos WHERE id = $1", id).Scan(&todo.ID, &todo.Title, &todo.Completed)
	if err != nil {
		if err == sql.ErrNoRows {
			return todo, errors.New("Todo not found")
		}
		return todo, err
	}

	return todo, nil
}

func (s *PostgresStorage) Add(todo models.Todo) (models.Todo, error) {
	err := s.db.QueryRow("INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id", todo.Title, todo.Completed).Scan(&todo.ID)
	if err != nil {
		return todo, err
	}

	return todo, nil
}

func (s *PostgresStorage) Update(todo models.Todo) error {
	_, err := s.db.Exec("UPDATE todos SET title = $1, completed = $2 WHERE id = $3", todo.Title, todo.Completed, todo.ID)
	return err
}

func (s *PostgresStorage) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM todos WHERE id = $1", id)
	return err
}
