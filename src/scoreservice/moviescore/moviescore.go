package moviescore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Operation string

const (
	OperationScore  Operation = "score"
	OperationSearch Operation = "search"
)

type (
	// Context represents the main application context
	Context struct {
		Provider  string
		Operation Operation
		Filename  string
		Query     string
		ID        string
	}

	// TODO: Make one result struct for both operations?
	Result struct {
		Provider  string       `json:"provider"`
		Operation Operation    `json:"operation"`
		Items     []ResultItem `json:"items"`
	}

	ResultItem struct {
		ID         string  `json:"id"`
		Score      float32 `json:"score"`
		ScoreClass string  `json:"score_class,omitempty"`
		Title      string  `json:"title,omitempty"`
		Poster     string  `json:"poster,omitempty"`
		Year       int     `json:"year,omitempty"`
	}

	// FindScoresResult result data structure of moviescore package operations
	FindScoresResult struct {
		IMDb   ResultItem
		Rotten ResultItem
	}

	// Provider is an interface used to reduce equal code
	Provider interface {
		Score(id string) (*Result, error)
		Search(query string) (*Result, error)
	}
)

var (
	ErrEmptyData = errors.New("empty data is not valid")
)

// UserAgents is a list of user agents that the scraper can use to trick the webmasters
// and hopefully don't get block
var UserAgents = [...]string{
	// Linus Firefox
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:43.0) Gecko/20100101 Firefox/43.0",
	// Mac Firefox
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.11; rv:43.0) Gecko/20100101 Firefox/43.0",
	// Mac Safari 4
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_2; de-at) AppleWebKit/531.21.8 (KHTML, like Gecko) Version/4.0.4 Safari/531.21.10",
	// Mac Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9",
	// Windows Chrome
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.125 Safari/537.36",
	// Windows IE 10
	"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0)",
	// Windows IE 11
	"Mozilla/5.0 (Windows NT 6.3; WOW64; Trident/7.0; rv:11.0) like Gecko",
	// Windows Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.10586",
	// Windows Firefox
	"Mozilla/5.0 (Windows NT 6.3; WOW64; rv:43.0) Gecko/20100101 Firefox/43.0",
	// iPhone
	"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B5110e Safari/601.1",
	// iPad
	"Mozilla/5.0 (iPad; CPU OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
	// Android
	"Mozilla/5.0 (Linux; Android 5.1.1; Nexus 7 Build/LMY47V) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.76 Safari/537.36",
}

// GetRandomUserAgent retrieves a random user agent
func GetRandomUserAgent() string {
	result := ""

	// Using current time nanosecond as seed
	seed := time.Now().Nanosecond()

	// Seed the random
	rand.Seed(int64(seed))

	// Get random user-agent
	size := len(UserAgents)
	result = UserAgents[rand.Int31n(int32(size))]

	return result
}

// Get performs a GET request to the given URL
func Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", GetRandomUserAgent())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		// TODO: better handle this
		return nil, errors.New("response code was not successfull")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// RunScore runs a score operation for the given provider and
// ID and outputs result temporary file with the given suffix
func RunScore(provider, id, suffix string) (*Result, error) {
	return run(provider, OperationScore, "", id, suffix)
}

// RunIMDbScore runs a score operation for IMDb with the given
// ID and outputs result temporary file with the given suffix
func RunIMDbScore(query, suffix string) (*Result, error) {
	return RunSearch(IMDB, query, suffix)
}

// RunRottenScore runs a score operation for Rotten with the given
// ID and outputs result temporary file with the given suffix
func RunRottenScore(id, suffix string) (*Result, error) {
	return RunScore(RottenT, id, suffix)
}

// RunSearch runs a search operation for the given provider and
// query and outputs result temporary file with the given suffix
func RunSearch(provider, query, suffix string) (*Result, error) {
	return run(provider, OperationSearch, query, "", suffix)
}

// RunIMDbSearch runs a search operation for Rotten with the given
// query and outputs result temporary file with the given suffix
func RunIMDbSearch(query, suffix string) (*Result, error) {
	return RunSearch(IMDB, query, suffix)
}

// RunRottenSearch runs a search operation for Rotten with the given
// query and outputs result temporary file with the given suffix
func RunRottenSearch(query, suffix string) (*Result, error) {
	return RunSearch(RottenT, query, suffix)
}

func run(provider string, operation Operation, query string, id string, suffix string) (*Result, error) {
	dir := os.TempDir()
	filename := filepath.Join(dir, "result_"+provider+"_"+suffix)

	ctx := &Context{
		Provider:  provider,
		Filename:  filename,
		Operation: operation,
		Query:     query,
		ID:        id,
	}

	err := ctx.run()
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close()
	os.Remove(f.Name())
	return parseOutput(contents)
}

func (ctx *Context) run() error {
	var p Provider
	var err error

	if ctx.Provider == IMDB {
		p = NewIMDb()
	} else if ctx.Provider == RottenT {
		p = NewRottenTomatoes()
	}

	if p != nil {
		var result *Result

		switch ctx.Operation {
		case OperationSearch:
			result, err = p.Search(ctx.Query)
		case OperationScore:
			result, err = p.Score(ctx.ID)
		default:
		}

		if result != nil {
			r := OutputFile(ctx.Filename, result)
			fmt.Printf("Outputted to: %s\n", r.Filename)
		}
	}

	return err
}

func parseOutput(output []byte) (*Result, error) {
	if output == nil || len(output) == 0 {
		return nil, ErrEmptyData
	}
	var result Result
	err := json.Unmarshal(output, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}
