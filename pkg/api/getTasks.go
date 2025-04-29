package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/DemArt555/go_final_project/pkg/models"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"` // Формат "20060102"
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func GetTasks(limit string, search string) ([]*Task, error) {
	//если параметр search пустой, возвращаем все задачи
	if search == "" {
		return GetAllTasks(limit)
	}
	//проверяем является ли search датой в формате 02.01.2006
	if date, err := time.Parse("02.01.2006", search); err == nil {
		// Если это дата, преобразуем ее в формат "20060102" и выполняем выборку по дате
		return GetTasksbyDate(date.Format("20060102"), limit)
	}
	//Если это не дата , выполняем поиск по подсроке в title и comment
	return GetTasksbySearch(search, limit)
}

// Функция получения всех задач
func GetAllTasks(limit string) ([]*Task, error) {

	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        ORDER BY date ASC 
        LIMIT ?
    `

	rows, err := db.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

// Функция получения задач по дате
func GetTasksbyDate(date string, limit string) ([]*Task, error) {

	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE date = ?
		ORDER BY date ASC 
        LIMIT ?
    `
	rows, err := db.DB.Query(query, date, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks by search: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return tasks, nil
}

type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}

// Хендлер получения задач
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {

	search := r.URL.Query().Get("search")
	tasks, err := GetTasks("50", search)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//Проверяем, что tasks не nil (чтобы в JSON был [], а не null)
	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	WriteJSON(w, TasksResp{Tasks: tasks})
}
func GetTasksbySearch(search string, limit string) ([]*Task, error) {
	//Приводим строку поиска к верхнему регистру и добавляем символы % для LIKE
	searchPattern := "%" + strings.ToUpper(search) + "%"

	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE UPPER(title) LIKE ? OR UPPER(comment) LIKE ?
		ORDER BY date ASC 
        LIMIT ?
    `
	rows, err := db.DB.Query(query, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks by search: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return tasks, nil
}

func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

//Получаем GetTask по id

func GetTaskById(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, err
	}
	return &task, nil
}

// Хэндлер для получения задачи по id
func HandleGetTaskById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteJSON(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := GetTaskById(id)
	if err != nil {
		WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, task)
}

// обновляем задачу
func UpdateTask(task *models.Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}

// хендлер для обновления задачи
func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Неверный формат данных"})
		return
	}

	// Проверка обязательных полей
	if task.ID == "" || task.Date == "" || task.Title == "" {
		WriteJSON(w, map[string]string{"error": "Не все обязательные поля заполнены"})
		return
	}

	// Проверка даты
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		WriteJSON(w, map[string]string{"error": "Неверный формат даты"})
		return
	}
	today := time.Now().Truncate(24 * time.Hour)
	parsedDate = parsedDate.Truncate(24 * time.Hour)
	if task.Repeat == "" && parsedDate.Before(today) {
		WriteJSON(w, map[string]string{"error": "Дата не может быть в прошлом"})
		return
	}

	// Проверка repeat
	if task.Repeat != "" {
		if !isValidRepeat(task.Repeat) {
			WriteJSON(w, map[string]string{"error": "Неверный формат repeat"})
			return
		}
	}

	err = UpdateTask(&task)
	if err != nil {
		WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, map[string]string{}) // пустой JSON
}

// Проверяем валидность правила повторения
func isValidRepeat(repeat string) bool {
	// допустим только формат "d N", где N — положительное число
	parts := strings.Split(repeat, " ")
	if len(parts) != 2 {
		return false
	}
	if parts[0] != "d" {
		return false
	}
	_, err := strconv.Atoi(parts[1])
	return err == nil
}
