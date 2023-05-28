package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/andrpech/ses-gen-tech/fs"
	"github.com/andrpech/ses-gen-tech/mail"
	"github.com/andrpech/ses-gen-tech/mw"
	"github.com/andrpech/ses-gen-tech/rate"
	"github.com/andrpech/ses-gen-tech/util"
)

const (
	port    string = ":8080"
	apiPath string = "/api"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

func main() {
	fmt.Println("Hello, gen tech!")

	e := echo.New()

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Config: %v", config)

	apiGroup := e.Group(apiPath)

	apiGroup.GET("/kenobi", healthCheck)
	apiGroup.GET("/rate", getRate)
	apiGroup.POST("/subscribe", subscribeEmail, mw.ParseFormData)
	apiGroup.POST("/sendEmails", sendEmails)

	fmt.Println("Server started")
	defer fmt.Println("Server stopped")

	e.Start(port)
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Hello there")
}

func getRate(c echo.Context) error {
	rate, err := rate.GetBinanceRate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rate)
}

func subscribeEmail(c echo.Context) error {
	email := c.Get("email").(string)

	emails, err := fs.ReadJson("db/emails.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Read file error: %v", err.Error())})
	}

	log.Printf("[subscribeEmail] emails: %v", emails)

	for _, e := range emails {
		if e.Email == email {
			return c.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Email '%v' already exists from '%v'.", e.Email, e.CreatedAt)})
		}
	}

	emails = append(emails, fs.Email{
		Email:     email,
		CreatedAt: time.Now().Format(time.RFC3339), // current time in RFC3339 format
	})

	err = fs.WriteJson("db/emails.json", emails)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Write file error: %v", err.Error())})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("E-mail '%v' додано.", email)})
}

func sendEmails(c echo.Context) error {
	rate, err := rate.GetBinanceRate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("[sendEmails] rate: %f", rate)

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	sender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

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

	json, err := fs.ReadJson("emails.json")
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
