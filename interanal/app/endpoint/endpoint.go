package endpoint

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/andrpech/ses-gen-tech/interanal/app/service/fs"
)

type Service interface {
	GetBinanceRate() (float64, error)
	ReadJson(filePath string) ([]fs.Email, error)
	WriteJson(filePath string, emails []fs.Email) error
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type Endpoint struct {
	s Service
}

func New(s Service) *Endpoint {
	return &Endpoint{
		s: s,
	}
}

//go:embed assets/email_template.html
var emailTemplate string

func (e *Endpoint) HealthCheck(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Hello there")
}

func (e *Endpoint) GetRate(ctx echo.Context) error {
	rate, err := e.s.GetBinanceRate()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, rate)
}

func (e *Endpoint) SubscribeEmail(ctx echo.Context) error {
	email := ctx.Get("email").(string)

	emails, err := e.s.ReadJson("db/emails.json")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Read file error: %v", err.Error())})
	}

	log.Printf("[SubscribeEmail] emails: %v", emails)

	for _, e := range emails {
		if e.Email == email {
			return ctx.JSON(http.StatusConflict, map[string]string{"error": fmt.Sprintf("Email '%v' already exists from '%v'.", e.Email, e.CreatedAt)})
		}
	}

	emails = append(emails, fs.Email{
		Email:     email,
		CreatedAt: time.Now().Format(time.RFC3339), // current time in RFC3339 format
	})

	err = e.s.WriteJson("db/emails.json", emails)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Write file error: %v", err.Error())})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("E-mail '%v' додано.", email)})
}

func (e *Endpoint) SendEmails(ctx echo.Context) error {
	rate, err := e.s.GetBinanceRate()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("[SendEmails] rate: %f", rate)

	subject := "A Galaxy Far, Far Away From The Current BTCUAH Exchange Rate"
	content := fmt.Sprintf(emailTemplate, "https://github.com/andrpech/ses-gen-tech/blob/main/assets/img/header.jpg?raw=true", rate, "https://github.com/andrpech/ses-gen-tech/blob/main/assets/img/footer.jpg?raw=true")

	json, err := e.s.ReadJson("db/emails.json")
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Read file error: %v", err.Error())})
	}

	log.Printf("[SendEmails] json: %v", json)

	// Create a new array without the CreatedAt field
	emails := make([]string, len(json))
	for i, email := range json {
		emails[i] = email.Email

		to := []string{emails[i]}

		err = e.s.SendEmail(subject, content, to, nil, nil, nil)
		if err != nil {
			log.Printf("[SubscribeEmail] failed to send email to %v: %v", to, err)
		} else {
			log.Printf("[SubscribeEmail] successfully sent email to %v", to)
		}
	}

	log.Printf("[SendEmails] emails: %v", emails)

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "E-mailʼи відправлено",
		"emails":  fmt.Sprintf("%v", emails),
	})
}
