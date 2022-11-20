package externalrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/kkr2/vessels/internal/config"
	"github.com/kkr2/vessels/internal/errors"
	"github.com/kkr2/vessels/internal/logger"
)

type WeatherClient interface {
	GetWeatherForDay(ctx context.Context, day time.Time) (float64, error) //beaufort
}

// Ths implementation of cache is not safe. Used as an example
// production env, key-value in memory db with TTL and eviction policy should be used

type kvcache struct {
	data map[string]float64
	mu   sync.Mutex
}

type weatherClient struct {
	cache  kvcache
	cfg    *config.Config
	logger logger.Logger
}

func NeweatherClient(cfg *config.Config, log logger.Logger) WeatherClient {
	return &weatherClient{
		cfg:    cfg,
		logger: log,
		cache: kvcache{
			data: make(map[string]float64),
		},
	}
}

func (wc *weatherClient) GetWeatherForDay(ctx context.Context, t time.Time) (float64, error) {
	operation := errors.Op("externalrpc.weatherRepository.GetWeatherForDay")
	dayStringFormat := fmt.Sprintf("%d-%02d-%02d", t.Year(), int(t.Month()), t.Day())
	wc.cache.mu.Lock()
	cachedRes, exists := wc.cache.data[dayStringFormat]
	wc.cache.mu.Unlock()
	if exists {
		return cachedRes, nil
	}
	res, err := wc.makeExternalCall(dayStringFormat)
	if err != nil {
		return 0, errors.E(operation, errors.KindExternalRPC, err)
	}
	// update cache
	wc.cache.mu.Lock()
	wc.cache.data[dayStringFormat] = res
	wc.cache.mu.Unlock()

	return res, nil

}

type ReqBody struct {
	Date string `json:"Date"`
}
type ResBody struct {
	Beaufort float64 `json:"Beaufort"`
}

func (wc *weatherClient) makeExternalCall(dayAsString string) (float64, error) {

	body := &ReqBody{
		Date: dayAsString,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(body)
	req, err := http.NewRequest("POST", wc.cfg.Server.WeatherApiUrl, payloadBuf)
	if err != nil {
		return 0, err
	}

	req.Header.Set("x-api-key", wc.cfg.Server.WeatherSecret)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	resbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var rb = new(ResBody)
	err = json.Unmarshal(resbody, &rb)
	if err != nil {
		return 0, err
	}

	return rb.Beaufort, nil
}
