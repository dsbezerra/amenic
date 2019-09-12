package moviescore

import (
	"encoding/json"
	"fmt"
	"os"
)

type (
	// OutputResult represents the output of the application, which is a
	// text file using JSON as its format.
	OutputResult struct {
		Filename string
		Data     string
	}
)

// OutputFile outputs struct data to a JSON file.
func OutputFile(filename string, data interface{}) *OutputResult {
	if filename == "" || data == nil {
		return nil
	}

	contents, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return nil
	}

	tmpfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0755)
	if _, err := tmpfile.Write(contents); err != nil {
		return nil
	}
	if err := tmpfile.Close(); err != nil {
		return nil
	}
	// NOTE: Don't forget to remove file from caller after process.
	// defer os.Remove(tmpfile.Name())
	return &OutputResult{
		Filename: tmpfile.Name(),
		Data:     string(contents),
	}
}
