package main

import (
	"Todolistick/handlers"
	"Todolistick/storage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Создаём роутер
	r := mux.NewRouter()

	// Подключаемся к БД
	dbStorage, err := storage.NewPostgresStorage("user=postgres password=kizoDB dbname=todolist sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer dbStorage.Close()

	// Создаём обработчики
	handler := handlers.TodoHandler{Storage: dbStorage}

	// Регистрируем маршруты
	r.HandleFunc("/todos", handler.GetAll).Methods("GET")
	r.HandleFunc("/todos/{id:[0-9]+}", handler.GetByID).Methods("GET")
	r.HandleFunc("/todos", handler.Add).Methods("POST")
	r.HandleFunc("/todos/{id:[0-9]+}", handler.Update).Methods("PUT")
	r.HandleFunc("/todos/{id:[0-9]+}", handler.Delete).Methods("DELETE")

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080")

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

}
