package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

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

type Email struct {
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

func main() {
	fmt.Println("Hello, gen tech!")

	e := echo.New()

	apiGroup := e.Group(apiPath)

	apiGroup.GET("/kenobi", healthCheck)
	apiGroup.GET("/rate", getRate)
	apiGroup.POST("/subscribe", subscribeEmail)

	fmt.Println("Server started")
	defer fmt.Println("Server stopped")

	e.Start(port)
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Hello there")
}

func getRate(c echo.Context) error {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbols=%v", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Api call error: %v", err.Error())})
	}
	defer resp.Body.Close()

	log.Println("Response status code:", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to parse response from Binance: %v", err.Error())})
	}

	// Check for error response
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Binance API error: %v", errResp.Msg)})
	}

	var rateResp RateResponse
	if err := json.Unmarshal(body, &rateResp); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": fmt.Sprintf("Failed to parse response from Binance: %v", err.Error()),
		})
	}

	log.Println(string(body))

	return c.JSONBlob(http.StatusOK, []byte(rateResp.Price))
}

func subscribeEmail(c echo.Context) error {
	email, err := parseFormData(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	emails, err := readJson("emails.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Read file error: %v", err.Error())})
	}

	log.Println(emails)

	for _, e := range emails {
		if e.Email == email {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Email already exists from %v", e.CreatedAt)})
		}
	}

	emails = append(emails, Email{
		Email:     email,
		CreatedAt: time.Now().Format(time.RFC3339), // current time in RFC3339 format
	})

	err = writeJson("emails.json", emails)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Write file error: %v", err.Error())})
	}

	return c.JSON(http.StatusOK, emails)
}

func parseFormData(c echo.Context) (string, error) {
	// Read the body into a string
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return "", errors.New("Invalid request")
	}
	bodyString := string(bodyBytes)

	// Parse form data
	formData, err := url.ParseQuery(bodyString)
	if err != nil {
		return "", errors.New("Invalid form data")
	}

	// Check for extra fields
	for key := range formData {
		if key != "email" {
			return "", errors.New(fmt.Sprintf("Unexpected form field: '%v'", key))
		}
	}

	// Get email
	email := formData.Get("email")
	if email == "" {
		return "", errors.New("Missing email field")
	}

	// Validate email
	if !isEmailValid(email) {
		return "", errors.New("Invalid email")
	}

	return email, nil
}

func isEmailValid(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match := regexp.MustCompile(emailRegex).MatchString
	return match(email)
}

func readJson(filePath string) ([]Email, error) {
	var emails []Email

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create file with empty array
		emptyArray := []Email{}
		jsonData, err := json.Marshal(emptyArray)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return nil, err
		}
		// Write  empty array to file
		err = os.WriteFile(filePath, jsonData, 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return nil, err
		}

		fmt.Printf("Created file '%s' with an empty array.\n", filePath)

		emails = emptyArray
	} else {
		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Read file content
		bytes, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON content into struct
		if err := json.Unmarshal(bytes, &emails); err != nil {
			return nil, err
		}
	}

	return emails, nil
}

func writeJson(filePath string, emails []Email) error {
	// Marshal the emails slice to JSON
	bytes, err := json.Marshal(emails)
	if err != nil {
		return err
	}

	// Write the JSON to the file
	err = os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
