package rate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

const symbol string = "BTCUAH"


type RateResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}


func GetBinanceRate () (float64, error ){
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%v", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Api call error: %v", err.Error()))
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to parse response from Binance: %v", err.Error()))
	}

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != 0 {
		return 0, errors.New(fmt.Sprintf("Binance API error: %v", errResp.Msg))
	}

	// Parse response
	var rateResp RateResponse
	if err := json.Unmarshal(body, &rateResp); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to parse response from Binance: %v", err.Error()))
	}

	log.Printf("[getBinanceRate] response.body: %v", string(body))

	// Convert string to integer
	num, err := strconv.ParseFloat(rateResp.Price, 64)
	if err != nil {
		fmt.Print("Error converting string to integer:", err)
		return 0, errors.New(fmt.Sprintf("Failed to convert rateResp.Price from string to integer: %v", err.Error()))
	}

	return num, nil
}