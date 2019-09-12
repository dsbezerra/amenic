package cloudinary

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

var (
	ImageChars       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ImagesNameLength = len(ImageChars) / 2
)

func generateFilename() string {
	result := strings.Builder{}
	rand.Seed(time.Now().Unix())
	for i := 0; i < ImagesNameLength; i++ {
		N := rand.Intn(ImagesNameLength)
		result.WriteByte(ImageChars[N])
	}
	return result.String()
}

// FileChecksum returns SHA1 file checksum
func FileChecksum(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return FileChecksumFromData(data)
}

// FileChecksumFromData ...
func FileChecksumFromData(data []byte) (string, error) {
	hash := sha1.New()
	io.WriteString(hash, string(data))
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
