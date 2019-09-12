package moviescore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// IMDB represents an abbreviation or short name for the Internet Movie Database website
const IMDB = "imdb"

const imdbBaseURL = "https://www.imdb.com"
const imdbAPIBaseURL = "https://v2.sg.media-imdb.com/suggests"

type (
	// IMDb represents an IMDB provider
	IMDb struct{}

	imdbSearchResult struct {
		V     int              `json:"v"`
		Query string           `json:"q"`
		Data  []imdbSearchItem `json:"d"`
	}

	imdbSearchItem struct {
		ID      string                `json:"id"`
		Image   []interface{}         `json:"i"`
		Label   string                `json:"l"`
		Q       string                `json:"q"`
		Subline string                `json:"s"`
		VT      int                   `json:"vt"`
		Videos  []imdbSearchItemVideo `json:"v"`
		Year    int                   `json:"y"`
	}

	imdbSearchItemVideo struct {
		Images  []interface{} `json:"i"`
		ID      string        `json:"id"`
		Label   string        `json:"l"`
		Subline string        `json:"s"`
	}

	imdbSearchOutputFormat struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Poster string `json:"poster"`
	}
)

// NewIMDb creates a new instance of IMDb provider
func NewIMDb() *IMDb {
	return &IMDb{}
}

// Search returns movies for a given query from IMDB suggests API
func (imdb *IMDb) Search(query string) (*Result, error) {
	if query == "" {
		return nil, nil
	}

	query = strings.Replace(query, " ", "_", -1)
	url := fmt.Sprintf("%s/%s/%s.json", imdbAPIBaseURL, string(query[0]), query)

	body, err := Get(url)
	if err != nil {
		return nil, err
	}

	str := string(body)
	start := strings.Index(str, "(") + 1
	end := strings.LastIndex(str, ")")
	if start > 0 && end > start {
		str = str[start:end]
	} else {
		return nil, errors.New("couldn't find string between `(` `)` tokens")
	}

	var searchResult imdbSearchResult
	err = json.Unmarshal([]byte(str), &searchResult)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Provider:  IMDB,
		Operation: OperationSearch,
	}

	items := make([]ResultItem, len(searchResult.Data))
	for i := 0; i < len(searchResult.Data); i++ {
		resource := searchResult.Data[i]
		item := ResultItem{
			ID:    resource.ID,
			Title: resource.Label,
			Year:  resource.Year,
		}

		if len(resource.Image) > 0 {
			item.Poster = getString(resource.Image[0])
		}

		items[i] = item
	}
	result.Items = items

	return result, err
}

// Score gets the score for the given imdb id
func (imdb *IMDb) Score(id string) (*Result, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}

	url := fmt.Sprintf("%s/title/%s", imdbBaseURL, id)
	body, err := Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	container := doc.Find("#title-overview-widget")
	scoreText := container.Find("div.ratings_wrapper > div.imdbRating > div.ratingValue > strong > span").Text()
	scoreText = strings.TrimSpace(scoreText)

	if scoreText == "" {
		return nil, fmt.Errorf("Couldn't find score for movie %s", id)
	}

	number, err := strconv.ParseFloat(scoreText, 32)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Provider:  IMDB,
		Operation: OperationScore,
		Items: []ResultItem{
			ResultItem{
				ID:    id,
				Score: float32(number),
			},
		},
	}
	return result, err
}

func getString(val interface{}) string {
	typ := reflect.TypeOf(val)
	if typ != nil && typ.Kind() == reflect.String {
		return reflect.ValueOf(val).String()
	}
	return ""
}
