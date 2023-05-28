package main

import (
	"log"

	"github.com/andrpech/ses-gen-tech/interanal/pkg/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

