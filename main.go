package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/DemArt555/go_final_project/pkg/api"
	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Инициализация базы данных
	if err := db.InitDb(); err != nil {
		fmt.Printf("Ошибка при инициализации БД: %v\n", err)
		return
	}
	//Получаем порт из переменной окружения TODO_port
	port := 7540 // порт по умолчанию
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		} else {
			fmt.Printf("Некорректный порт в переменной TODO_PORT: %s. Используется порт по умолчанию %d\n", envPort, port)
		}
	}

	// Настройка роутера
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	// Маршруты
	r.Get("/api/nextdate", api.NextDateHandler)
	r.Post("/api/task", api.AddTaskHandler)
	r.Get("/api/tasks", api.GetTasksHandler)
	r.Get("/api/task", api.HandleGetTaskById)
	r.Put("/api/task", api.HandleUpdateTask)
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Chi!"))
	})
	r.Post("/api/task/done", api.HandleTaskDone) //завершение задачи
	r.Delete("/api/task", api.HandleDeleteTask)  //удаление задачи
	// Статические файлы
	r.Handle("/", http.FileServer(http.Dir("web")))

	// Запуск сервера
	fmt.Printf("Starting server on :%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
	}
}
