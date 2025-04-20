package main

import (
	"Todolistick/handlers"
	"Todolistick/storage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Подключение к PostgreSQL
	connStr := "user=postgres password=kizoDB dbname=todolist host=localhost sslmode=disable"
	dbStorage, err := storage.NewPostgresStorage(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbStorage.Close()

	// Логируем все записи из базы данных (для отладки)
	todos, err := dbStorage.GetAll()
	if err != nil {
		log.Printf("Failed to get all todos: %v", err)
	} else {
		log.Printf("Current todos in database: %v", todos)
	}

	// Создаём роутер и хендлер
	r := mux.NewRouter()
	handler := handlers.TodoHandler{Storage: dbStorage}

	// Регистрируем маршруты
	r.HandleFunc("/todos", handler.GetAll).Methods("GET")
	r.HandleFunc("/todos/{id}", handler.GetByID).Methods("GET")
	r.HandleFunc("/todos", handler.Add).Methods("POST")
	r.HandleFunc("/todos/{id}", handler.Update).Methods("PUT")
	r.HandleFunc("/todos/{id}", handler.Delete).Methods("DELETE")

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
