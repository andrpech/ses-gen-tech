package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	port    string = ":8080"
	apiPath string = "/api"
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

func main() {
	fmt.Println("Hello, gen tech!")

	e := echo.New()

	apiGroup := e.Group(apiPath)

	apiGroup.GET("/kenobi", healthCheck)
	apiGroup.GET("/rate", getRate)


	fmt.Println("Server started")
	defer fmt.Println("Server stopped")

	e.Start(port)
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Hello there")
}

func getRate(c echo.Context) error {
	url := `https://api.binance.com/api/v3/ticker/price?symbol=` + symbol

	resp, err := http.Get(url)
	if err != nil {
		return c.String(http.StatusBadRequest, `Error occurred: `+err.Error())
	}
	defer resp.Body.Close()

	log.Println("Response status code:", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, `Error occurred: `+err.Error())
	}

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": `Error from Binance API: ` + errResp.Msg})
	}

	var rateResp RateResponse
	if err := json.Unmarshal(body, &rateResp); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": `Failed to parse response from Binance: ` + err.Error(),
		})
	}

	log.Println(string(body))

	return c.JSONBlob(http.StatusOK, []byte(rateResp.Price))
}


