package fileutil

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
)

var errFilepath = errors.New("invalid filepath")

// TextFileHandler is a handler to read a text file.
type TextFileHandler struct {
	Filepath   string
	File       *os.File
	Scanner    *bufio.Scanner
	LineNumber uint
}

// StartFile is responsible to open the given file and setup the handler.
func (h *TextFileHandler) StartFile(filepath string) error {
	if filepath == "" {
		return errFilepath
	}

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	h.Filepath = filepath
	h.File = file
	h.Scanner = bufio.NewScanner(file)
	h.LineNumber = 0

	return nil
}

// ConsumeNextLine consumes the next line in the buffer and skips blank ones.
// If a non-blank line is found, the line is returned as string, otherwise it keeps
// consuming until reach a non-blank line or the end of the buffer.
func (h *TextFileHandler) ConsumeNextLine() (string, bool) {
	for h.Scanner.Scan() {
		line := h.Scanner.Text()

		h.LineNumber++

		lineStr := string(line)
		if lineStr == "" {
			continue
		}

		return lineStr, true
	}

	return "", false
}

// CloseFile closes the text file handler.
func (h *TextFileHandler) CloseFile() {
	err := h.File.Close()
	if err != nil {
		fmt.Printf("Couldn't close file %s! reason: %s\n", h.Filepath, err.Error())
	}
	h.File = nil
	h.LineNumber = 0
}

// PrintError is a print error routine that provides the current line number information.
func (h *TextFileHandler) PrintError(format string, args ...interface{}) {
	fullFormat := "Error at line %d: " + format
	var completeArgs []interface{}
	completeArgs = append(completeArgs, h.LineNumber)
	completeArgs = append(completeArgs, args...)
	log.Printf(fullFormat, completeArgs...)
}
