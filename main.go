package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/jordan-wright/email"
)

const (
	port    string = ":8080"
	apiPath string = "/api"
)

const symbol string = "BTCUAH"

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

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

type Config struct {
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func main() {
	fmt.Println("Hello, gen tech!")

	e := echo.New()

	config, err := LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Config: %v", config)

	apiGroup := e.Group(apiPath)

	apiGroup.GET("/kenobi", healthCheck)
	apiGroup.GET("/rate", getRate)
	apiGroup.POST("/subscribe", subscribeEmail)
	apiGroup.POST("/sendEmails", sendEmails)

	fmt.Println("Server started")
	defer fmt.Println("Server stopped")

	e.Start(port)
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Hello there")
}

func getRate(c echo.Context) error {
	rate, err := getBinanceRate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, rate)
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

	log.Printf("[subscribeEmail] emails: %v", emails)

	for _, e := range emails {
		if e.Email == email {
			return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Email '%v' already exists from '%v'.", e.Email, e.CreatedAt)})
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

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("E-mail '%v' додано.", email)})
}

func sendEmails(c echo.Context) error {
	rate, err := getBinanceRate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("[sendEmails] rate: %f", rate)

	config, err := LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A Galaxy Far, Far Away From The Current BTCUAH Exchange Rate"
	text := `
	<h1>Hello there</h1>
	<p>I trust thi
	s message finds you well amidst the bluster of the cosmos. I am writing to you from the dunes of Tatooine, under the twin suns where much is uncertain but the shifting sands.
	
	In the boundless expanse of the universe, one constant has always intrigued me - the fluctuating dance of currency exchange rates. A dance that, in many ways, resembles the delicate balance between the light and dark sides of the Force.
	
	At present, the Bitcoin (BTC) to Ukrainian Hryvnia (UAH) exchange rate stands as such:
	
	BTC: 1
	UAH: %f
	
	Just as the Force flows through and around us, so too does the rhythm of finance, reminding us of the interconnectedness of all things. Like the murmurs of the midichlorians, these values whisper stories of the world economy's ebb and flow, the strength of currencies, and the dynamics of the crypto market.
	
	Though it might seem as remote and indifferent as the binary suns of Tatooine, I encourage you to embrace this knowledge as a Jedi would -- with curiosity, caution, and a readiness to learn.
	
	The ever-shifting exchange rate may seem like a wretched hive of scum and volatility, but I am confident that with vigilance and wisdom, we can navigate its intricacies just as surely as a seasoned pilot navigates an asteroid field.
	
	I shall endeavor to keep you updated on this interstellar journey through the realm of financial constellations. May the Force, and the wisdom to harness its essence, be with you always.
	
	Your humble servant in the Force,
	
	Ben Kenobi</a></p>
	`
	content := fmt.Sprintf(text, rate)

	json, err := readJson("emails.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Read file error: %v", err.Error())})
	}

	log.Printf("[sendEmails] json: %v", json)

	// Create a new array without the CreatedAt field
	emails := make([]string, len(json))
	for i, email := range json {
		emails[i] = email.Email

		to := []string{emails[i]}

		err = sender.SendEmail(subject, content, to, nil, nil, nil)
		if err != nil {
			log.Printf("[subscribeEmail] failed to send email to %v: %v", to, err)
		} else {
			log.Printf("[subscribeEmail] successfully sent email to %v", to)
		}
	}

	log.Printf("[sendEmails] emails: %v", emails)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "E-mailʼи відправлено",
		"emails":  fmt.Sprintf("%v", emails),
	})
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

	return strings.ToLower(email), nil
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

func getBinanceRate () (float64, error ){
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

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}