package wiki

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mbydanov/tg_golang_bot/internal/models"
)

type SearchResultsExt struct {
	models.SearchResults
}

// Запись данных в структуры
func (sr *SearchResultsExt) UnmarshalJSON(bs []byte) error {
	array := []interface{}{}
	if err := json.Unmarshal(bs, &array); err != nil {
		return err
	}
	sr.Query = array[0].(string)
	for i := range array[1].([]interface{}) {
		sr.Results = append(sr.Results, models.Result{
			Name:        array[1].([]interface{})[i].(string),
			Description: array[2].([]interface{})[i].(string),
			Url:         array[3].([]interface{})[i].(string),
		})
	}
	return nil
}

// Получение и отправка данных
func WikipediaAPI(request string) (answer []string) {
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
		sr := &models.SearchResults{}
		if err = json.Unmarshal([]byte(contents), sr); err != nil {
			s[0] = "Something going wrong, try to change your question"
		}

		//Проверяем не пустая ли наша структура
		if !sr.Ready {
			s[0] = "Something going wrong, try to change your question"
		}

		//Проходим через нашу структуру и отправляем данные в срез с ответом
		for i := range sr.Results {
			s[i] = sr.Results[i].Url
		}
	}
	return s
}

// Ввиду того что мы отправляем URL нам нужно конвертировать сообщение от пользователя в часть URL.
// Зачем это нужно, затем что пользователь может отправить боту не одно а два слова через пробел,
// нам же нужно заменить пробел так чтобы он стал частью URL. Этим займется функция urlEncoded.
// Конвертируем запрос для использование в качестве части URL
func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func WikipediaGET(searchText string) (answer []string) {
	// Устанавливаем язык для поиска в википедии
	language := os.Getenv("LANGUAGE")

	//Создаем url для поиска
	ms, _ := UrlEncoded(searchText)

	url := ms
	request := "https://" + language + ".wikipedia.org/w/api.php?action=opensearch&search=" + url + "&limit=3&origin=*&format=json"
	message := WikipediaAPI(request)
	return message
}
