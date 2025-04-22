package storage

import (
	"Todolistick/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

// TodoStore defines methods for working with Todo records.
type TodoStore interface {
	GetAll(ctx context.Context) ([]models.Todo, error)
	GetByID(ctx context.Context, id int) (models.Todo, error)
	Add(ctx context.Context, todo models.Todo) (models.Todo, error)
	Update(ctx context.Context, todo models.Todo) error
	Delete(ctx context.Context, id int) error
	Close() error
}

// PostgresStorage implements TodoStore using PostgreSQL.
type PostgresStorage struct {
	db          *sql.DB
	stmtGetAll  *sql.Stmt
	stmtGetByID *sql.Stmt
	stmtAdd     *sql.Stmt
	stmtUpdate  *sql.Stmt
	stmtDelete  *sql.Stmt
}

// NewPostgresStorage opens a connection and prepares statements.
func NewPostgresStorage(connStr string) (TodoStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Create table if not exists
	mig := `
	CREATE TABLE IF NOT EXISTS todos (
	    id SERIAL PRIMARY KEY,
	    title TEXT NOT NULL,
	    completed BOOLEAN NOT NULL DEFAULT false
	);
	`
	if _, err = db.Exec(mig); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate todos table: %w", err)
	}

	s := &PostgresStorage{db: db}
	ctx := context.Background()

	// Prepare statements
	if s.stmtGetAll, err = db.PrepareContext(ctx, "SELECT id, title, completed FROM todos"); err != nil {
		return nil, fmt.Errorf("prepare GetAll: %w", err)
	}
	if s.stmtGetByID, err = db.PrepareContext(ctx, "SELECT id, title, completed FROM todos WHERE id = $1"); err != nil {
		return nil, fmt.Errorf("prepare GetByID: %w", err)
	}
	if s.stmtAdd, err = db.PrepareContext(ctx, "INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id"); err != nil {
		return nil, fmt.Errorf("prepare Add: %w", err)
	}
	if s.stmtUpdate, err = db.PrepareContext(ctx, "UPDATE todos SET title = $1, completed = $2 WHERE id = $3"); err != nil {
		return nil, fmt.Errorf("prepare Update: %w", err)
	}
	if s.stmtDelete, err = db.PrepareContext(ctx, "DELETE FROM todos WHERE id = $1"); err != nil {
		return nil, fmt.Errorf("prepare Delete: %w", err)
	}

	return s, nil
}

// Close closes all statements and the DB.
func (s *PostgresStorage) Close() error {
	if s.stmtGetAll != nil {
		s.stmtGetAll.Close()
	}
	if s.stmtGetByID != nil {
		s.stmtGetByID.Close()
	}
	if s.stmtAdd != nil {
		s.stmtAdd.Close()
	}
	if s.stmtUpdate != nil {
		s.stmtUpdate.Close()
	}
	if s.stmtDelete != nil {
		s.stmtDelete.Close()
	}
	return s.db.Close()
}

// GetAll returns all todos.
func (s *PostgresStorage) GetAll(ctx context.Context) ([]models.Todo, error) {
	rows, err := s.stmtGetAll.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("query GetAll: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			return nil, fmt.Errorf("scan GetAll: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows GetAll: %w", err)
	}
	return todos, nil
}

// GetByID returns a todo by ID.
func (s *PostgresStorage) GetByID(ctx context.Context, id int) (models.Todo, error) {
	var t models.Todo
	err := s.stmtGetByID.QueryRowContext(ctx, id).Scan(&t.ID, &t.Title, &t.Completed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return t, errors.New("todo not found")
		}
		return t, fmt.Errorf("query GetByID: %w", err)
	}
	return t, nil
}

// Add inserts a new todo and returns it.
func (s *PostgresStorage) Add(ctx context.Context, todo models.Todo) (models.Todo, error) {
	err := s.stmtAdd.QueryRowContext(ctx, todo.Title, todo.Completed).Scan(&todo.ID)
	if err != nil {
		return todo, fmt.Errorf("query Add: %w", err)
	}
	return todo, nil
}

// Update modifies an existing todo.
func (s *PostgresStorage) Update(ctx context.Context, todo models.Todo) error {
	res, err := s.stmtUpdate.ExecContext(ctx, todo.Title, todo.Completed, todo.ID)
	if err != nil {
		return fmt.Errorf("exec Update: %w", err)
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return errors.New("todo not found")
	}
	return nil
}

// Delete removes a todo by ID.
func (s *PostgresStorage) Delete(ctx context.Context, id int) error {
	res, err := s.stmtDelete.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("exec Delete: %w", err)
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return errors.New("todo not found")
	}
	return nil
}
