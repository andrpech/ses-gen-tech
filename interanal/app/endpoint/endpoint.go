package endpoint

import (
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
	text := `
	<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            font-family: Arial, sans-serif;
        }
        .content {
            margin: 0 auto;
            max-width: 600px;
        }
        h1, h2 {
            color: #02569B;
        }
    </style>
</head>
<body>
    <div class="content">
				<img align="center" alt="Header image" src="%s" width="600" style="max-width:1200px;padding-bottom:0;display:inline!important;vertical-align:bottom;border:0;height:auto;outline:none;text-decoration:none">
        <h1>Hello there</h1>
        <p>
            I trust this message finds you well amidst the bluster of the cosmos. I am writing to you from the dunes of Tatooine, under the twin suns where much is uncertain but the shifting sands.
        </p>
				<p>
						In the boundless expanse of the universe, one constant has always intrigued me - the fluctuating dance of currency exchange rates. A dance that, in many ways, resembles the delicate balance between the light and dark sides of the Force.
				</p>
        <b>At present, the Bitcoin (BTC) to Ukrainian Hryvnia (UAH) exchange rate:</b>
        <p>
					<b>BTC:<b> 1<br>
					<b>UAH:<b> %f
        </p>
				<p>
						Just as the Force flows through and around us, so too does the rhythm of finance, reminding us of the interconnectedness of all things. Like the murmurs of the midichlorians, these values whisper stories of the world economy's ebb and flow, the strength of currencies, and the dynamics of the crypto market.
				</p>
				<p>
						Though it might seem as remote and indifferent as the binary suns of Tatooine, I encourage you to embrace this knowledge as a Jedi would -- with curiosity, caution, and a readiness to learn.
				</p>
				<p>
						The ever-shifting exchange rate may seem like a wretched hive of scum and volatility, but I am confident that with vigilance and wisdom, we can navigate its intricacies just as surely as a seasoned pilot navigates an asteroid field.
				</p>
        <p>
            I shall endeavor to keep you updated on this interstellar journey through the realm of financial constellations. May the Force, and the wisdom to harness its essence, be with you always.
        </p>
        <p>
            Your humble servant in the Force,<br>
            Ben Kenobi
        </p>
				<img align="center" alt="Footer image" src="%s" width="600" style="max-width:1200px;padding-bottom:0;display:inline!important;vertical-align:bottom;border:0;height:auto;outline:none;text-decoration:none">
    </div>
</body>
</html>`
	content := fmt.Sprintf(text, "https://github.com/andrpech/ses-gen-tech/blob/main/assets/img/header.jpg?raw=true", rate, "https://github.com/andrpech/ses-gen-tech/blob/main/assets/img/footer.jpg?raw=true")

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
