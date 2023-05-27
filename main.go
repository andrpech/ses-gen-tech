package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

const port string = ":8080"


func main() {
	fmt.Println("Hello, gen tech!")

	s := echo.New()

	s.GET("/kenobi", func(c echo.Context) error { return c.String(200, "Hello there") })

	fmt.Println("Server started")
	defer fmt.Println("Server stopped")

	s.Start(port)
}
