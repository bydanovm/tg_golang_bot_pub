//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	_ "github.com/lib/pq"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type searchResults struct {
	ready   bool
	query   string
	results []result
}

type result struct {
	name, description, url string
}

// Запись данных в структуры
func (sr *searchResults) UnmarshalJSON(bs []byte) error {
	array := []interface{}{}
	if err := json.Unmarshal(bs, &array); err != nil {
		return err
	}
	sr.query = array[0].(string)
	for i := range array[1].([]interface{}) {
		sr.results = append(sr.results, result{
			array[1].([]interface{})[i].(string),
			array[2].([]interface{})[i].(string),
			array[3].([]interface{})[i].(string),
		})
	}
	return nil
}

// Получение и отправка данных
func wikipediaAPI(request string) (answer []string) {
	//Создаем срез на 3 элемента
	s := make([]string, 3)

	//Отправляем запрос
	if response, err := http.Get(request); err != nil {
		s[0] = "Wikipedia is not respond"
	} else {
		defer response.Body.Close()

		//Считываем ответ
		contents, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		//Отправляем данные в структуру
		sr := &searchResults{}
		if err = json.Unmarshal([]byte(contents), sr); err != nil {
			s[0] = "Something going wrong, try to change your question"
		}

		//Проверяем не пустая ли наша структура
		if !sr.ready {
			s[0] = "Something going wrong, try to change your question"
		}

		//Проходим через нашу структуру и отправляем данные в срез с ответом
		for i := range sr.results {
			s[i] = sr.results[i].url
		}
	}
	return s
}

// Ввиду того что мы отправляем URL нам нужно конвертировать сообщение от пользователя в часть URL.
// Зачем это нужно, затем что пользователь может отправить боту не одно а два слова через пробел,
// нам же нужно заменить пробел так чтобы он стал частью URL. Этим займется функция urlEncoded.
// Конвертируем запрос для использование в качестве части URL
func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Теперь нам нужно как то взаимодействовать с БД.
// Создаем переменные в которых мы будем хранить данные переменных окружения для подключению к БД.
var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var user = os.Getenv("USER")
var password = os.Getenv("PASSWORD")
var dbname = os.Getenv("DBNAME")
var sslmode = os.Getenv("SSLMODE")

var dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

// Создаем таблицу users в БД при подключении к ней
func createTable() error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаем таблицу users
	if _, err = db.Exec(`CREATE TABLE users(ID SERIAL PRIMARY KEY, TIMESTAMP TIMESTAMP DEFAULT CURRENT_TIMESTAMP, USERNAME TEXT, CHAT_ID INT, MESSAGE TEXT, ANSWER TEXT);`); err != nil {
		return err
	}
	return nil
}

// Таблицу мы создали, и нам нужно заносить в нее данные, этим займется следующая функция.
// Собираем данные полученные ботом
func collectData(username string, chatid int64, message string, answer []string) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Конвертируем срез с ответом в строку
	answ := strings.Join(answer, ", ")

	//Создаем SQL запрос
	data := `INSERT INTO users(username, chat_id, message, answer) VALUES($1, $2, $3, $4);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, `@`+username, chatid, message, answ); err != nil {
		return err
	}

	return nil
}

// Также давайте напишем функцию которая будет считать количество уникальных пользователей
// которые писали боту, чтобы отдавать это число пользователям если они отправят боту нужную команду.
func getNumberOfUsers() (int64, error) {
	var count int64

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT COUNT(DISTINCT username) FROM users;")
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Создаем бота
func telegramBot() {
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a wikipedia bot, i can search information in a wikipedia, send me something what you want find in Wikipedia.")
				bot.Send(msg)
			case "/number_of_users":
				if os.Getenv("DB_SWITCH") == "on" {
					// Присваиваем количество пользоватьелей использовавших бота в num переменную
					num, err := getNumberOfUsers()
					if err != nil {
						//Отправлем сообщение
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error.")
						bot.Send(msg)
					}

					// Создаем строку которая содержит колличество пользователей использовавших бота
					ans := fmt.Sprintf("%d peoples used me for search information in Wikipedia", num)

					// Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, ans)
					bot.Send(msg)
				} else {
					// Отправлем сообщение
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database not connected, so i can't say you how many peoples used me.")
					bot.Send(msg)
				}
			default:
				// Устанавливаем язык для поиска в википедии
				language := os.Getenv("LANGUAGE")

				//Создаем url для поиска
				ms, _ := urlEncoded(update.Message.Text)

				url := ms
				request := "https://" + language + ".wikipedia.org/w/api.php?action=opensearch&search=" + url + "&limit=3&origin=*&format=json"

				// Присваем данные среза с ответом в переменную message
				message := wikipediaAPI(request)

				if os.Getenv("DB_SWITCH") == "on" {
					// Отправляем username, chat_id, message, answer в БД
					if err := collectData(update.Message.Chat.UserName, update.Message.Chat.ID, update.Message.Text, message); err != nil {

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
func main() {

	time.Sleep(10 * time.Second)

	// Создаем таблицу
	if os.Getenv("CREATE_TABLE") == "yes" {
		if os.Getenv("DB_SWITCH") == "on" {
			if err := createTable(); err != nil {
				panic(err)
			}
		}
	}

	time.Sleep(10 * time.Second)

	// Вызываем бота
	telegramBot()
}
