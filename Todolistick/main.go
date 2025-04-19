package main

import (
	"Todolistick/handlers"
	"Todolistick/storage"
	"github.com/gorilla/mux"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	// Создаём новый роутер
	r := mux.NewRouter()

	// Подключение к базе данных
	dbStorage, err := storage.NewSQLiteStorage("todolist.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStorage.Close() // Закрытие подключения к базе данных при завершении работы

	// Логируем все записи из базы данных (для отладки)
	todos, err := dbStorage.GetAll()
	if err != nil {
		log.Printf("Failed to get all todos: %v", err)
	} else {
		log.Printf("Current todos in database: %v", todos)
	}

	// Создаём хендлер
	handler := handlers.TodoHandler{Storage: dbStorage}

	// Регистрируем маршруты
	r.HandleFunc("/todos", handler.GetAll).Methods("GET")
	r.HandleFunc("/todos/{id}", handler.GetByID).Methods("GET")
	r.HandleFunc("/todos", handler.Add).Methods("POST")
	r.HandleFunc("/todos/{id}", handler.Update).Methods("PUT")
	r.HandleFunc("/todos/{id}", handler.Delete).Methods("DELETE")

	// Логируем запуск сервера
	log.Println("Server is running on port 8080")

	// Запускаем сервер с нашим роутером
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
