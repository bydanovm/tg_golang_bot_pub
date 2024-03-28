package main

import (
	"os"
	"time"

	"github.com/mbydanov/tg_golang_bot/internal/database"
	retrievercoins "github.com/mbydanov/tg_golang_bot/internal/retrieverCoins"
	"github.com/mbydanov/tg_golang_bot/internal/tgbot"
)

func main() {

	time.Sleep(5 * time.Second)

	// Создаем таблицу
	if os.Getenv("CREATE_TABLE") == "yes" {
		if os.Getenv("DB_SWITCH") == "on" {
			if err := database.CreateTables(); err != nil {
				panic(err)
			}
		}
	}

	time.Sleep(2 * time.Second)

	// Вызов функции автоматического обновления КВ
	go retrievercoins.RunRetrieverCoins(300)
	// Вызываем бота
	tgbot.TelegramBot()
}
