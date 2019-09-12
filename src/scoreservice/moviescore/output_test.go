package moviescore

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestOutputFormat struct {
	FieldOne   int      `json:"field_one"`
	FieldTwo   string   `json:"field_two"`
	FieldThree []string `json:"field_three"`
}

func TestOutputFile(t *testing.T) {
	data := TestOutputFormat{
		FieldOne:   1,
		FieldTwo:   "two",
		FieldThree: []string{"one", "two", "three"},
	}
	result := OutputFile("test_file", data)
	assert.NotNil(t, result)
	defer os.Remove(result.Filename)
	contents, err := ioutil.ReadFile(result.Filename)
	assert.NoError(t, err)
	assert.Equal(t, string(contents), result.Data)
}
