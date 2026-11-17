package main

import (
	"log"
	"net/http"
	"os"
	"src/internals/database"
	"src/internals/handlers"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://taskuser:taskpass@localhost:5432/tasksdb?sslmode=disable"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("Начинаем запуск сервера на порту %s", serverPort)

	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	log.Println("Подключение к базе данных успешно")

	taskStore := database.NewTaskStore(db)

	handler := handlers.NewHandler(taskStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", methodHandler(handler.GetAllTasks, http.MethodGet))
	mux.HandleFunc("/tasks/create", methodHandler(handler.CreateTask, http.MethodPost))
	mux.HandleFunc("/tasks/", TaskIDHandler(handler))

	loggedMux := loggingMiddleware(mux)

	serverAddr := ":" + serverPort

	err = http.ListenAndServe(serverAddr, loggedMux)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func methodHandler(handlerFunc http.HandlerFunc, allowedMethods string) http.HandlerFunc {
	return func(wr http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethods {
			http.Error(wr, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}
		handlerFunc(wr, r)
}
}

func TaskIDHandler(handler *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTask(w, r)
		case http.MethodPut:
			handler.UpdateTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}