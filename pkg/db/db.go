package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/DemArt555/go_final_project/pkg/models"
	_ "modernc.org/sqlite"
)

const DbFile = "scheduler.db"

// SQL схема базы данных
const Schema = `
CREATE TABLE IF NOT EXISTS scheduler 
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date VARCHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128)
);

CREATE INDEX IF NOT EXISTS IDX_date ON scheduler (date);
`

var DB *sql.DB

//Подключение к БД и создание файла, если он не существует

func InitDb() error {
	//Проверка, существует ли файл БД "scheduler.db"
	if _, err := os.Stat(DbFile); os.IsNotExist(err) {
		fmt.Println("Файл базы данных не найден. Создаем новый:", DbFile)
		file, err := os.Create(DbFile)
		if err != nil {
			return fmt.Errorf("ошибка создания файла БД: %v", err)
		}
		file.Close()
	}

	// Подключение к ДБ
	var err error
	DB, err = sql.Open("sqlite", DbFile)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}
	// Применение Schema
	if _, err := DB.Exec(Schema); err != nil {
		return fmt.Errorf("ошибка применения схемы: %v", err)

	}

	return nil
}

// Функция добавления задачи
func AddTask(task *models.Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// UpdateDate обновляет дату выполнения задачи в базе данных
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, next, id)
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

// DeleteTask удаляет задачу из базы данных по её id
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
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
