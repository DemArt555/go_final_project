package db

import (
	"database/sql"
	"fmt"
	"os"

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
