package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"EWallet/pkg/metrics"

	"github.com/sirupsen/logrus"
)

var ErrCurrencyNotFound = errors.New("err currency not found")

type Rate struct {
	log    *logrus.Entry
	xrHost string
	apiKey string
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
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewExchangeRate(log *logrus.Logger, xrHost string, apiKey string) *Rate {
	return &Rate{
		log:    log.WithField("component", "exchange"),
		xrHost: xrHost,
		apiKey: apiKey,
	}
}

func (e *Rate) GetRate(ctx context.Context, currency string, amount float64) (float64, error) {
	started := time.Now()
	defer func() {
		metrics.MetricHTTPRequestDuration.Observe(time.Since(started).Seconds())
	}()
	amountStr := fmt.Sprintf("%v", amount)
	url := e.xrHost + currency + "&from=rub&amount=" + amountStr
	fmt.Println(url)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("apikey", e.apiKey)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		return 1.0, fmt.Errorf("exchange api internal srver error: %w", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return 0, fmt.Errorf("%s: %w", currency, ErrCurrencyNotFound)
	default:
		metrics.MetricErrCount.WithLabelValues("GetRate").Inc()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return 0, fmt.Errorf("err handling another error (unexpected status code: %d),fail to read response body: %w", res.StatusCode, err)
		}
		return 0, fmt.Errorf("unexpected status code: %d body: %s", res.StatusCode, string(body))
	}
	var result Resp
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("err decoding response: %w", err)
	}
	return result.Result, nil
}
