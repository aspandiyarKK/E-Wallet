package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type ExchangeRate struct {
	log *logrus.Entry
}
type Resp struct {
	Success bool `json:"success"`
	Query   struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	} `json:"query"`
	Info struct {
		Timestamp int     `json:"timestamp"`
		Rate      float64 `json:"rate"`
	} `json:"info"`
	Date   string  `json:"date"`
	Result float64 `json:"result"`
}

func NewExchange(log *logrus.Logger) *ExchangeRate {
	return &ExchangeRate{
		log: log.WithField("component", "exchange"),
	}
}

func (e *ExchangeRate) GetRate(ctx context.Context, currency string, amount float64) (float64, error) {
	amountStr := fmt.Sprintf("%v", amount)
	url := "https://api.apilayer.com/exchangerates_data/convert?to=" + currency + "&from=rub&amount=" + amountStr
	fmt.Println(url)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("apikey", "fcMPeBqh49hXoGziH4US5YqMbEA4njMi")

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("converting error: %w", err)
	}
	var s Resp
	err = json.Unmarshal(body, &s)
	if err != nil {
		return 0, fmt.Errorf("invalid input: %w", err)
	}
	return s.Result, nil
}
