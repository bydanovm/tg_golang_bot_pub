package coinmarketcup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mbydanov/tg_golang_bot/internal/database"
	"github.com/mitchellh/mapstructure"
)

func GetLatest(cryptocurrencies string) (answer []string) {
	s := make([]string, 0)
	// Проверяем наличие криптовалюты в БД
	fields := database.DictCrypto{}
	expLst := []database.Expressions{
		database.Expressions{database.CryptoName, database.EQ, cryptocurrencies},
	}
	rs, find, err := database.ReadDataRow(&fields, expLst, 1)
	if err != nil {
		s = append(s, "Возвращена ошибка при поиске в БД: "+err.Error())
		return s
	}
	// Если запись найдена, возвращаем из БД
	// Нужна проверка на время обновления (возможно SQL запросом)
	if find {
		mapstructure.Decode(rs[0], &fields)
		str := fmt.Sprintf("Криптовалюта: %s\nЦена: %.3f %s\nОбновлено: %s",
			fields.CryptoName,
			fields.CryptoLastPrice,
			"USD",
			fields.CryptoUpdate.Format("2006-01-02 15:04:05"),
		)
		s = append(s, str)
		return s
	}
	// Если записи в БД нет или время обновления истекло, то вызываем API
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest", nil)
	if err != nil {
		s = append(s, "Возвращена ошибка: "+err.Error())
	}

	q := url.Values{}
	q.Add("symbol", cryptocurrencies)
	q.Add("convert", "USD")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", os.Getenv("API_CMC"))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		s = append(s, "Возвращена ошибка: "+err.Error())
	}
	respBody, _ := io.ReadAll(resp.Body)
	qla := &QuotesLatestAnswer{}
	if err = json.Unmarshal([]byte(respBody), qla); err != nil {
		s = append(s, "Возвращена ошибка: "+err.Error())
	}
	if qla.Error_code != 0 {
		s = append(s, "Возвращена ошибка: "+qla.Error_message)
	}
	for i := range qla.QuotesLatestAnswerResults {
		str := fmt.Sprintf("Криптовалюта: %s\nЦена: %.3f %s\nОбновлено: %s",
			qla.QuotesLatestAnswerResults[i].Symbol,
			qla.QuotesLatestAnswerResults[i].Price,
			qla.QuotesLatestAnswerResults[i].Currency,
			qla.QuotesLatestAnswerResults[i].Last_updated.Format("2006-01-02 15:04:05"),
		)
		s = append(s, str)
	}
	return s
}

func (qla *QuotesLatestAnswer) UnmarshalJSON(bs []byte) error {
	var quotesLatest QuotesLatest
	if err := json.Unmarshal(bs, &quotesLatest); err != nil {
		return err
	}
	qla.Error_code = quotesLatest.Status.ErrorCode
	qla.Error_message = quotesLatest.Status.Error_message
	for _, value0 := range quotesLatest.Data {
		qla.QuotesLatestAnswerResults = append(qla.QuotesLatestAnswerResults, QuotesLatestAnswerResult{
			Id:           value0[0].Id,
			Name:         value0[0].Name,
			Symbol:       value0[0].Symbol,
			Cmc_rank:     value0[0].Cmc_rank,
			Price:        value0[0].Quote["USD"].Price,
			Currency:     "USD",
			Last_updated: value0[0].Quote["USD"].Last_updated,
		})
		cryptoprices := map[string]string{
			"CryptoId":     fmt.Sprintf("%v", value0[0].Id),
			"CryptoPrice":  fmt.Sprintf("%v", value0[0].Quote["USD"].Price),
			"CryptoUpdate": fmt.Sprint(value0[0].Quote["USD"].Last_updated.Format("2006-01-02 15:04:05")),
		}
		dictCryptos := map[string]string{
			"CryptoId":        fmt.Sprintf("%v", value0[0].Id),
			"CryptoName":      fmt.Sprintf("%v", value0[0].Symbol),
			"CryptoLastPrice": fmt.Sprintf("%v", value0[0].Quote["USD"].Price),
			"CryptoUpdate":    fmt.Sprint(value0[0].Quote["USD"].Last_updated.Format("2006-01-02 15:04:05")),
		}
		if err := database.WriteData("DictCrypto", dictCryptos); err != nil {
			return err
		}
		if err := database.WriteData("cryptoprices", cryptoprices); err != nil {
			return err
		}
	}
	return nil
}
