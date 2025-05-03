package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/DemArt555/go_final_project/pkg/api"
	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	// Инициализация базы данных
	if err := db.InitDb(); err != nil {
		fmt.Printf("Ошибка при инициализации БД: %v\n", err)
		return
	}

	//Настройка роутера
	r := chi.NewRouter()
	//Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	//установка правильной интерпретации пути к FrontEnd у сервера
	fs := http.FileServer(http.Dir("web"))
	r.Handle("/*", http.StripPrefix("/", fs))
	//Маршруты
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

	// Значение порта по-умолчанию
	defaultPort := 7540

	// Получаем значение переменной окружения TODO-PORT
	portStr := os.Getenv("TODO_PORT")

	// Пробуем конвертировать в int
	port := defaultPort
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		} else {
			fmt.Printf("Неверное значение TODO_PORT: %s. Используется порт по умолчанию: %d\n", portStr, defaultPort)
		}
	}

	fmt.Printf("Starting server on :%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
	}
}
