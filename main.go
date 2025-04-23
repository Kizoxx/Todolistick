package main

import (
	"Todolistick/handlers"
	"Todolistick/storage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	// Создаём роутер
	r := mux.NewRouter()

	// Подключаемся к БД
	dbStorage, err := storage.NewPostgresStorage()
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

	// Применяем middleware Logger ко всем маршрутам
	loggedRouter := Logger(r)

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080")
	err = http.ListenAndServe(":8080", loggedRouter)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] Completed in %v", time.Since(start))
	})
}
