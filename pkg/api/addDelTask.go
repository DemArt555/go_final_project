package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/DemArt555/go_final_project/pkg/models"
)

// обработчик добавления задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		WriteJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		WriteJSONError(w, "Title is required", http.StatusBadRequest)
		return

	}

	if err := CheckDate(&task); err != nil {
		log.Printf("Date validation failed: %v", err)
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return

	}

	id, err := db.AddTask(&task)
	if err != nil {
		log.Printf("Database error: %v", err)
		WriteJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]int64{"id": id})
}

// проверяем дату для обработчика добавления задачи
func CheckDate(task *models.Task) error {
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	if task.Repeat == "d 1" {
		task.Date = now.Format("20060102")
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		return errors.New("invalid date format")
	}

	if afterNow(parsedDate, now) {
		return nil
	}

	if task.Repeat == "d 1" || task.Repeat == "" {
		task.Date = now.Format("20060102")
	} else {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		task.Date = nextDate
	}

	return nil
}

// HandleTaskDone обрабатывает POST-запрос /api/task/done
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем id из параметров запроса
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJSONError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	// Получаем задачу из базы данных
	task, err := GetTaskById(id)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Ошибка получения задачи: %v", err), http.StatusBadRequest)
		return
	}

	// Если задача одноразовая (repeat пустое), удаляем её
	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			WriteJSONError(w, fmt.Sprintf("Ошибка удаления задачи: %v", err), http.StatusInternalServerError)
			return
		}
		WriteJSON(w, map[string]string{}) // Успешное удаление
		return
	}

	// Если задача периодическая, вычисляем следующую дату
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Ошибка вычисления следующей даты: %v", err), http.StatusBadRequest)
		return
	}

	// Обновляем дату в базе данных
	err = db.UpdateDate(nextDate, id)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Ошибка обновления даты: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON в случае успеха
	WriteJSON(w, map[string]string{})
}

// HandleDeleteTask обрабатывает DELETE-запрос /api/task
func HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodDelete {
		WriteJSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем id из параметров запроса
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJSONError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	// Удаляем задачу
	err := db.DeleteTask(id)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Ошибка удаления задачи: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON в случае успеха
	WriteJSON(w, map[string]string{})
}
