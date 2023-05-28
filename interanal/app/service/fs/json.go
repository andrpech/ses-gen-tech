package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Email struct {
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type FileService interface {
	ReadJson(filePath string) ([]Email, error)
	WriteJson(filePath string, emails []Email) error
}

type FS struct{}

func NewFS() FileService {
	return &FS{}
}

func (fs *FS) ReadJson(filePath string) ([]Email, error) {
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

func (fs *FS) WriteJson(filePath string, emails []Email) error {
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
