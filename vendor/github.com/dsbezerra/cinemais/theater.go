package cinemais

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/cinemais/util"
	"github.com/gocolly/colly"
)

const (
	// Selector of the container that contains theater list.
	scTheaters = "#indexContainer > div.selectSpotProgramacao ul"
)

// Theater ...
type Theater struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	City City   `json:"city"`
}

// Theaters ...
type Theaters []Theater

// GetTheaters ...
func GetTheaters() (Theaters, error) {
	return New().GetTheaters()
}

// GetTheaters ...
func (c *Cinemais) GetTheaters() (Theaters, error) {
	var result Theaters
	var err error

	c.collector.OnHTML(scTheaters, func(e *colly.HTMLElement) {
		// The theater list is within a UL element.
		if e.Name == "ul" {
			e.DOM.Find("li").EachWithBreak(func(i int, s *goquery.Selection) bool {
				t, err := parseTheater(s)
				if err != nil {
					return false
				}
				result.AppendUniq(t)
				return true
			})
		}
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/programacao", c.BaseURL.String())
	c.collector.Visit(url)

	return result, err
}

// IsEqual checks whether two city instances are equal or not.
func (t *Theater) IsEqual(theater Theater) bool {
	return t.ID == theater.ID && t.Name == theater.Name && t.City.IsEqual(theater.City)
}

// IsValid checks whether the city instance is valid or not.
func (t *Theater) IsValid() bool {
	return t.ID != 0 && t.Name != "" && t.City.IsValid()
}

// AppendUniq ...
func (t *Theaters) AppendUniq(theater *Theater) {
	exists := false

	for _, e := range *t {
		if e.IsEqual(*theater) {
			exists = true
			break
		}
	}

	if !exists {
		*t = append(*t, *theater)
	}
}

func parseTheater(s *goquery.Selection) (*Theater, error) {
	var result *Theater
	var err error

	// Get theater id from id attribute
	val, exists := s.Attr("id")
	if !exists {
		log.Println("theater id is missing")
		err = ErrNotFound
		return nil, err
	}

	// Convert to integer
	ID, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("couldn't convert '%s' to integer", val)
		return nil, err
	}

	result = &Theater{ID: ID}

	text := util.GetTextTrimmed(s.Contents().Not("script"))
	textParts := strings.Split(text, " - ")
	size := len(textParts)
	if size > 0 {
		result.Name = fmt.Sprintf("Cinemais %s", textParts[0])
	}

	if size == 2 {
		result.City = City{
			Name:           textParts[0],
			FederativeUnit: textParts[1],
		}
	} else if size == 3 {
		result.City = City{
			Name:           textParts[1],
			FederativeUnit: textParts[2],
		}
	}

	if result == nil || !result.IsValid() {
		log.Printf("couldn't get theater information from text '%s'\n", text)
		err = ErrNotFound
	}

	return result, err
}
