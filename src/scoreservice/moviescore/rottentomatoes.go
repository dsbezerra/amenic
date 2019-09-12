package moviescore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// RottenT represents an abbreviation or short name for the RottenTomatoes website
const RottenT = "rotten"

const rottenBaseURL = "https://www.rottentomatoes.com"
const rottenAPIBaseURL = rottenBaseURL + "/napi/"

const scoreClassRotten = "rotten"
const scoreClassFresh = "fresh"
const scoreClassCertifiedFresh = "certified_fresh"

type (
	// RottenTomatoes represents an IMDB provider
	RottenTomatoes struct{}

	/* Response struct for url:
	   https://www.rottentomatoes.com/napi/search?query="something"
	*/
	rtSearchResult struct {
		ActorCount     int           `json:"actorCount"`
		Actors         []rtActor     `json:"actors"`
		CriticCount    int           `json:"criticCount"`
		Critics        []rtCritic    `json:"critics"`
		FranchiseCount int           `json:"franchiseCount"`
		Franchises     []rtFranchise `json:"franchises"`
		MovieCount     int           `json:"movieCount"`
		Movies         []rtMovie     `json:"movies"`
		TvCount        int           `json:"tvCount"`
		TvSeries       []rtTvShow    `json:"tvSeries"`
	}

	rtActor struct {
		Image string `json:"image"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	}

	rtCritic struct {
		Image        string   `json:"image"`
		Name         string   `json:"name"`
		Publications []string `json:"publications"`
		URL          string   `json:"url"`
	}

	rtFranchise struct {
		Image string `json:"image"`
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	rtMovie struct {
		CastItems  []rtCastItem `json:"castItems"`
		Image      string       `json:"image"`
		MeterClass string       `json:"meterClass"`
		MeterScore uint         `json:"meterScore"`
		Name       string       `json:"name"`
		Subline    string       `json:"subline"`
		URL        string       `json:"url"`
		Year       int          `json:"year"`
	}

	rtCastItem struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	rtTvShow struct {
		Image      string `json:"image"`
		MeterClass string `json:"meterClass"`
		MeterScore uint   `json:"meterScore"`
		StartYear  uint   `json:"startYear"`
		EndYear    uint   `json:"endYear"`
		Title      string `json:"title"`
		URL        string `json:"url"`
	}

	rtSearchOutputFormat struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Poster     string `json:"poster"`
		MeterClass string `json:"meterClass"`
		MeterScore uint   `json:"meterScore"`
	}
)

// NewRottenTomatoes creates a new instance of RottenTomatoes provider
func NewRottenTomatoes() *RottenTomatoes {
	return &RottenTomatoes{}
}

// Search for movie, actors, shows, franchises, etc, using rotten public api.
func (rt *RottenTomatoes) Search(query string) (*Result, error) {
	if query == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%ssearch/?limit=5&query=%s", rottenAPIBaseURL, url.QueryEscape(query))
	body, err := Get(url)
	if err != nil {
		return nil, err
	}

	var searchResult rtSearchResult
	err = json.Unmarshal(body, &searchResult)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Provider:  RottenT,
		Operation: OperationSearch,
	}

	items := make([]ResultItem, len(searchResult.Movies))
	for i := 0; i < len(searchResult.Movies); i++ {
		movie := searchResult.Movies[i]
		items[i] = ResultItem{
			ID:         movie.URL,
			Title:      movie.Name,
			Poster:     movie.Image,
			Score:      float32(movie.MeterScore),
			ScoreClass: movie.MeterClass,
			Year:       movie.Year,
		}
	}
	result.Items = items

	return result, err
}

// Score gets the score for the given rotten page path as id
func (rt *RottenTomatoes) Score(id string) (*Result, error) {
	if id == "" {
		return nil, nil
	}

	path := ensurePathHasM(id)
	body, err := Get(fmt.Sprintf("%s%s", rottenBaseURL, path))
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	root := doc.Find("#topSection > div.col-sm-17.col-xs-24.score-panel-wrap")

	var title, meterClass string
	var meterScore uint

	title = strings.TrimSpace(root.Find("div.mop-ratings-wrap.score_panel > h1").Text())
	meterScore = scoreAsInt(root.Find("#tomato_meter_link > span.mop-ratings-wrap__percentage").Text())
	icon := root.Find("span.meter-tomato.icon")
	val, exists := icon.Attr("class")
	if exists {
		if strings.Contains(val, scoreClassRotten) {
			meterClass = scoreClassRotten
		} else if strings.Contains(val, scoreClassFresh) {
			meterClass = scoreClassFresh
		} else if strings.Contains(val, scoreClassCertifiedFresh) {
			meterClass = scoreClassCertifiedFresh
		}
	}

	item := ResultItem{
		ID:         path,
		Title:      title,
		Score:      float32(meterScore),
		ScoreClass: meterClass,
	}
	result := &Result{
		Provider:  RottenT,
		Operation: OperationScore,
		Items:     []ResultItem{item},
	}
	return result, err
}

func ensurePathHasM(path string) string {
	if strings.HasPrefix(path, "/m/") {
		return path
	} else if strings.HasPrefix(path, "m/") {
		path = "/" + path
	} else {
		if path[0] == '/' {
			path = "/m" + path
		} else {
			path = "/m/" + path
		}
	}
	return path
}

func scoreAsInt(scoreText string) uint {
	var result uint
	trimmed := strings.TrimFunc(scoreText, func(r rune) bool {
		return !unicode.IsNumber(r)
	})

	if trimmed != "" {
		number, err := strconv.Atoi(trimmed)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		result = uint(number)
	}

	return result
}
