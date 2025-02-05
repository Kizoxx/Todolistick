package main

import (
	"Todolistick/handlers"
	"Todolistick/storage" // Исправлен регистр: "Storage" → "storage"
	"github.com/gorilla/mux"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

func main() {
	r := mux.NewRouter()

	// Подключение к базе данных
	dbStorage, err := storage.NewSQLiteStorage("todolist.db")
	if err != nil {
		log.Fatal(err)
	}
	defer dbStorage.Close() // Закрытие подключения к базе данных при завершении работы

	handler := Handlers.TodoHandler{Storage: dbStorage}

	r.HandleFunc("/todos", handler.GetAll).Methods("GET")
	r.HandleFunc("/todos/{id}", handler.GetByID).Methods("GET")
	r.HandleFunc("/todos", handler.Add).Methods("POST")
	r.HandleFunc("/todos/{id}", handler.Update).Methods("PUT")
	r.HandleFunc("/todos/{id}", handler.Delete).Methods("DELETE")

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
