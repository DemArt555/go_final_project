package main

import (
	"fmt"

	"github.com/DemArt555/go_final_project/pkg/db"
	"github.com/DemArt555/go_final_project/pkg/server"
)

func main() {
	// Инициализация базы данных
	if err := db.InitDb(); err != nil {
		fmt.Printf("Ошибка при инициализации БД: %v\n", err)
		return
	}

	//Запускаем сервер

	if err := server.Run(); err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
	}

}
