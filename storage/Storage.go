package storage

import (
	"Todolistick/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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
func NewPostgresStorage() (TodoStore, error) {
	// Формируем строку подключения из переменных окружения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// Добавляем отладочный вывод для проверки строки подключения
	log.Printf("Попытка подключения: %s", connStr)

	// Открываем соединение
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}
		// Проверяем, готова ли база данных
		if err = db.Ping(); err == nil {
			log.Println("Успешное подключение к базе данных!")
			break
		}
		log.Printf("База данных не готова, ждем... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	// Создаем таблицу, если она не существует
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

	// Подготавливаем запросы
	if s.stmtGetAll, err = db.PrepareContext(ctx, "SELECT id, title, completed FROM todos"); err != nil {
		db.Close()
		return nil, fmt.Errorf("prepare GetAll: %w", err)
	}
	if s.stmtGetByID, err = db.PrepareContext(ctx, "SELECT id, title, completed FROM todos WHERE id = $1"); err != nil {
		db.Close()
		return nil, fmt.Errorf("prepare GetByID: %w", err)
	}
	if s.stmtAdd, err = db.PrepareContext(ctx, "INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id"); err != nil {
		db.Close()
		return nil, fmt.Errorf("prepare Add: %w", err)
	}
	if s.stmtUpdate, err = db.PrepareContext(ctx, "UPDATE todos SET title = $1, completed = $2 WHERE id = $3"); err != nil {
		db.Close()
		return nil, fmt.Errorf("prepare Update: %w", err)
	}
	if s.stmtDelete, err = db.PrepareContext(ctx, "DELETE FROM todos WHERE id = $1"); err != nil {
		db.Close()
		return nil, fmt.Errorf("prepare Delete: %w", err)
	}

	return s, nil
}

// Close closes all statements and the DB.
func (s *PostgresStorage) Close() error {
	var errs []error
	if s.stmtGetAll != nil {
		if err := s.stmtGetAll.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stmtGetAll: %w", err))
		}
		s.stmtGetAll = nil
	}
	if s.stmtGetByID != nil {
		if err := s.stmtGetByID.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stmtGetByID: %w", err))
		}
		s.stmtGetByID = nil
	}
	if s.stmtAdd != nil {
		if err := s.stmtAdd.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stmtAdd: %w", err))
		}
		s.stmtAdd = nil
	}
	if s.stmtUpdate != nil {
		if err := s.stmtUpdate.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stmtUpdate: %w", err))
		}
		s.stmtUpdate = nil
	}
	if s.stmtDelete != nil {
		if err := s.stmtDelete.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close stmtDelete: %w", err))
		}
		s.stmtDelete = nil
	}
	if err := s.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close db: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors while closing: %v", errs)
	}
	return nil
}

// GetAll returns all todos.
func (s *PostgresStorage) GetAll(ctx context.Context) ([]models.Todo, error) {
	log.Println("Получение всех задач")
	rows, err := s.stmtGetAll.QueryContext(ctx)
	if err != nil {
		log.Printf("Ошибка получения всех задач: %v", err)
		return nil, fmt.Errorf("query GetAll: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			log.Printf("Ошибка сканирования задачи: %v", err)
			return nil, fmt.Errorf("scan GetAll: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка обработки строк: %v", err)
		return nil, fmt.Errorf("rows GetAll: %w", err)
	}
	log.Printf("Успешно получено %d задач", len(todos))
	return todos, nil
}

// GetByID returns a todo by ID.
func (s *PostgresStorage) GetByID(ctx context.Context, id int) (models.Todo, error) {
	log.Printf("Получение задачи с ID: %d", id)
	var t models.Todo
	err := s.stmtGetByID.QueryRowContext(ctx, id).Scan(&t.ID, &t.Title, &t.Completed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Задача с ID %d не найдена", id)
			return t, errors.New("todo not found")
		}
		log.Printf("Ошибка получения задачи с ID %d: %v", id, err)
		return t, fmt.Errorf("query GetByID: %w", err)
	}
	log.Printf("Успешно получена задача: %v", t)
	return t, nil
}

// Add inserts a new todo and returns it.
func (s *PostgresStorage) Add(ctx context.Context, todo models.Todo) (models.Todo, error) {
	log.Printf("Добавление задачи: %v", todo)
	err := s.stmtAdd.QueryRowContext(ctx, todo.Title, todo.Completed).Scan(&todo.ID)
	if err != nil {
		log.Printf("Ошибка добавления задачи: %v", err)
		return todo, fmt.Errorf("query Add: %w", err)
	}
	log.Printf("Успешно добавлена задача с ID: %d", todo.ID)
	return todo, nil
}

// Update modifies an existing todo.
func (s *PostgresStorage) Update(ctx context.Context, todo models.Todo) error {
	log.Printf("Обновление задачи: %v", todo)
	res, err := s.stmtUpdate.ExecContext(ctx, todo.Title, todo.Completed, todo.ID)
	if err != nil {
		log.Printf("Ошибка обновления задачи: %v", err)
		return fmt.Errorf("exec Update: %w", err)
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		log.Printf("Задача с ID %d не найдена для обновления", todo.ID)
		return errors.New("todo not found")
	}
	log.Println("Успешно обновлена задача")
	return nil
}

// Delete removes a todo by ID.
func (s *PostgresStorage) Delete(ctx context.Context, id int) error {
	log.Printf("Удаление задачи с ID: %d", id)
	res, err := s.stmtDelete.ExecContext(ctx, id)
	if err != nil {
		log.Printf("Ошибка удаления задачи: %v", err)
		return fmt.Errorf("exec Delete: %w", err)
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		log.Printf("Задача с ID %d не найдена для удаления", id)
		return errors.New("todo not found")
	}
	log.Println("Успешно удалена задача")
	return nil
}
