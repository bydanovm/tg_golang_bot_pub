package tgbot

import (
	"fmt"
	"os"
	"reflect"

	_ "github.com/lib/pq"
	"github.com/mbydanov/tg_golang_bot/internal/coinmarketcup"
	"github.com/mbydanov/tg_golang_bot/internal/database"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

// Создаем бота
func TelegramBot() {
	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	// Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получаем обновления от бота
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Проверяем что от пользователья пришло именно текстовое сообщение
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			switch update.Message.Text {
			case "/start":
				// Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a bot.")
				bot.Send(msg)
			case "/number_of_users":
				if os.Getenv("DB_SWITCH") == "on" {
					// Присваиваем количество пользователей использовавших бота в num переменную
					num, err := database.GetNumberOfUsers()
					if err != nil {
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error.")
						bot.Send(msg)
					}

					// Создаем строку которая содержит колличество пользователей использовавших бота
					ans := fmt.Sprintf("%d peoples used me", num)

					// Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, ans)
					bot.Send(msg)
				} else {
					// Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database not connected, so i can't say you how many peoples used me.")
					bot.Send(msg)
				}
			default:
				message := coinmarketcup.GetLatest(update.Message.Text)

				if os.Getenv("DB_SWITCH") == "on" {
					// Отправляем username, chat_id, message, answer в БД
					if err := database.CollectData(update.Message.Chat.UserName, update.Message.Chat.ID, update.Message.Text, message); err != nil {

						// Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error, but bot still working.")
						bot.Send(msg)
					}
				}

				// Проходим через срез и отправляем каждый элемент пользователю
				for _, val := range message {

					// Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, val)
					bot.Send(msg)
				}
			}
		} else {
			// Отправлем сообщение
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use the words for search.")
			bot.Send(msg)
		}
	}
}
