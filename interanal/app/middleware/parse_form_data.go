package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/andrpech/ses-gen-tech/tools"
)

func ParseFormData(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Read the body into a string
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		}
		bodyString := string(bodyBytes)

		// Parse form data
		formData, err := url.ParseQuery(bodyString)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid form data."})
		}

		// Check for extra fields
		for key := range formData {
			if key != "email" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Unexpected form field: '%v'.", key)})
			}
		}

		// Get email
		email := formData.Get("email")
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing 'email' field. Please enter your email address."})
		}

		// Validate email
		if !tools.IsEmailValid(email) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid email. Please enter a valid email address."})
		}

		email = strings.ToLower(email)

		c.Set("email", email)

		return next(c)
	}
}
