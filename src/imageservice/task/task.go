package task

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dsbezerra/amenic/src/lib/persistence"
)

var (
	ErrNoImage          = errors.New("no image file")
	ErrImageNotModified = errors.New("image not modified")
	ErrInvalidInput     = errors.New("invalid input")
)

// RunAll executes rountine tasks for this imageservice microservice
func RunAll(data persistence.DataAccessLayer) {
	EnsureImagesChecksum(data)
}

func fileChecksum(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return dataChecksum(data)
}

func dataChecksum(data []byte) (string, error) {
	if len(data) == 0 {
		return "", ErrInvalidInput
	}
	return fmt.Sprintf("%x", sha1.Sum(data)), nil
}

func isImageFile(contentType string) bool {
	return strings.HasPrefix(contentType, "image")
}
