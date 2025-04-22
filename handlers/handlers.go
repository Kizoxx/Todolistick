package handlers

import (
	"Todolistick/models"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// TodoStorage определяет интерфейс для работы с хранилищем тудушек.
type TodoStorage interface {
	GetAll(ctx context.Context) ([]models.Todo, error)
	GetByID(ctx context.Context, id int) (models.Todo, error)
	Add(ctx context.Context, todo models.Todo) (models.Todo, error)
	Update(ctx context.Context, todo models.Todo) error
	Delete(ctx context.Context, id int) error
}

// TodoHandler обрабатывает HTTP-запросы и использует хранилище.
type TodoHandler struct {
	Storage TodoStorage
}

// Получить все туду
func (h *TodoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	todos, err := h.Storage.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(todos)
}

// Получить туду по ID
func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	todo, err := h.Storage.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(todo)
}

// Добавить новую туду
func (h *TodoHandler) Add(w http.ResponseWriter, r *http.Request) {
	var todo models.Todo
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addedTodo, err := h.Storage.Add(r.Context(), todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(addedTodo)
}

// Обновить существующую туду
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todo models.Todo
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if todo.ID != 0 && todo.ID != id {
		http.Error(w, "ID in URL does not match ID in body", http.StatusBadRequest)
		return
	}

	if todo.ID == 0 {
		todo.ID = id
	}

	err = h.Storage.Update(r.Context(), todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(todo)
}

// Удалить туду по ID
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.Storage.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
