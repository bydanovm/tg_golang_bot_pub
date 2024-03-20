package coinmarketcup

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func GetLatest(cryptocurrencies string) (answer []string) {
	s := make([]string, 0)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
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
		str := fmt.Sprintf("Криптовалюта: %s\nЦена: %.3f %s",
			qla.QuotesLatestAnswerResults[i].Symbol,
			qla.QuotesLatestAnswerResults[i].Price,
			qla.QuotesLatestAnswerResults[i].Currency)
		s = append(s, str)
	}
	return s
}

func (qla *QuotesLatestAnswer) UnmarshalJSON(bs []byte) error {
	var quotesLatest QuotesLatest
	// array := []interface{}{}
	if err := json.Unmarshal(bs, &quotesLatest); err != nil {
		return err
	}
	qla.Error_code = quotesLatest.Status.ErrorCode
	qla.Error_message = quotesLatest.Status.Error_message
	for _, value0 := range quotesLatest.Data {
		qla.QuotesLatestAnswerResults = append(qla.QuotesLatestAnswerResults, QuotesLatestAnswerResult{
			Id:       value0[0].Id,
			Name:     value0[0].Name,
			Symbol:   value0[0].Symbol,
			Cmc_rank: value0[0].Cmc_rank,
			Price:    value0[0].Quote["USD"].Price,
		})
	}
	return nil
}
