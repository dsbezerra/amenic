package fileutil

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

// Struct2Json Generates a JSON file with the given name and interface
func Struct2Json(filename string, v interface{}) ([]byte, error) {
	if filename == "" {
		return nil, errors.New("invalid filename")
	}

	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filename, data, os.ModePerm)
	return data, err
}
