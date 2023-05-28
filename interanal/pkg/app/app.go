package app

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"

	"github.com/andrpech/ses-gen-tech/interanal/app/endpoint"
	"github.com/andrpech/ses-gen-tech/interanal/app/middleware"
	"github.com/andrpech/ses-gen-tech/interanal/app/service"
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

type App struct {
	endpoint *endpoint.Endpoint
	service  *service.Service
	echo     *echo.Echo
}

func New() (*App, error) {

	app := &App{}

	app.service = service.New()
	app.endpoint = endpoint.New(app.service)

	fmt.Println("[New] hello, gen tech!")

	app.echo = echo.New()

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Config: %v", config)

	apiGroup := app.echo.Group(apiPath)

	apiGroup.GET("/kenobi", app.endpoint.HealthCheck)
	apiGroup.GET("/rate", app.endpoint.GetRate)
	apiGroup.POST("/subscribe", app.endpoint.SubscribeEmail, middleware.ParseFormData)
	apiGroup.POST("/sendEmails", app.endpoint.SendEmails)

	return app, nil
}

func (app *App) Run() error {
	fmt.Println("[Run] server started")

	err := app.echo.Start(port)
	if err != nil {
		log.Fatal(err)
	}
	defer fmt.Println("[Run] server stopped")

	return nil
}
