package main

import (
	"os"
	"time"

	"github.com/mbydanov/tg_golang_bot/internal/database"
	"github.com/mbydanov/tg_golang_bot/internal/tgbot"
)

func main() {

	time.Sleep(10 * time.Second)

	// Создаем таблицу
	if os.Getenv("CREATE_TABLE") == "yes" {
		if os.Getenv("DB_SWITCH") == "on" {
			if err := database.CreateTable(); err != nil {
				panic(err)
			}
		}
	}

	time.Sleep(10 * time.Second)

	// Вызываем бота
	tgbot.TelegramBot()
}
