package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/DemArt555/go_final_project/pkg/models"
)

// функция добавления задачи
func AddTask(task *models.Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// обработчик добавления задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJsonAddTask(w, map[string]string{"error": "Invalid JSON format"})
		return
	}

	if task.Title == "" {
		writeJsonAddTask(w, map[string]string{"error": "Title is required"})
		return
	}

	if err := CheckDate(&task); err != nil {
		writeJsonAddTask(w, map[string]string{"error": err.Error()})
		return
	}

	id, err := AddTask(&task)
	if err != nil {
		writeJsonAddTask(w, map[string]string{"error": "Database error"})
		return
	}

	writeJsonAddTask(w, map[string]int64{"id": id})
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

// обрабатываем Json для addtask
func writeJsonAddTask(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

// DeleteTask удаляет задачу из базы данных по её id
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача с id %s не найдена", id)
	}

	return nil
}

// UpdateDate обновляет дату выполнения задачи в базе данных
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := db.DB.Exec(query, next, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления даты: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновления: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача с id %s не найдена", id)
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
		err = DeleteTask(id)
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
	err = UpdateDate(nextDate, id)
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
	err := DeleteTask(id)
	if err != nil {
		WriteJSONError(w, fmt.Sprintf("Ошибка удаления задачи: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем пустой JSON в случае успеха
	WriteJSON(w, map[string]string{})
}
