package handlers

import (
	"encoding/json"
	"net/http"
	"src/internals/database"
	"src/internals/models"
	"strconv"
	"strings"
)


type Handler struct {
	store *database.TaskStore
}


func NewHandler(store *database.TaskStore) *Handler {
	return &Handler{store: store}
}

func respondWithJSON(w http.ResponseWriter, statuscode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)

	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, statuscode int, message string) {
	respondWithJSON(w, statuscode, map[string]string{"error": message})
}

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	var completed *bool

	if param := r.URL.Query().Get("completed"); param != "" {
		value, err := strconv.ParseBool(param)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Некорректное значение параметра 'completed'!")
			return
		}
		completed = &value
	}

	tasks, err := h.store.GetAll(completed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Произошла ошибка при получении задач!")
		return
	}
	respondWithJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный ID задачи!")
		return
	}

	task, err := h.store.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input_task models.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input_task);
	err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректно отправлены данные для создания задачи!")
		return
	}

	if strings.TrimSpace(input_task.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Заголовок задачи должен быть заполнен!")
		return
	}

	task, err := h.store.Create(input_task)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/")
	idStr := pathParts[0]


	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный ID задачи!")
		return
	}

	var input_task models.UpdateTaskInput

	if err := json.NewDecoder(r.Body).Decode(&input_task); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректно отправлены данные для обновления задачи!")
		return
	}

	if input_task.Title != nil && strings.TrimSpace(*input_task.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Заголовок задачи должен быть заполнен!")
		return
	}

	task, err := h.store.Update(id, input_task)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			respondWithError(w, http.StatusNotFound, "Задача не найдена!")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/")
	idStr := pathParts[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный ID задачи!")
		return
	}

	err = h.store.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			respondWithError(w, http.StatusNotFound, "Задача не найдена!")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Задача успешно удалена!"})
}
