package cinemais

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/cinemais/util"
	"github.com/gocolly/colly"
)

//
// TODO: complete
//
// SELECTORS VARIABLES NAME CONVENTION
//
// All variables prefixed with a lowercased 's' is a variable
// holding a selector string.
//
// If variable name is followed by:
//
// c - This selectors refers to container of desired scrape region.
//
//
// The rest is the name used to identify them in code.
//

const (
	// Selector of the container that contains city list.
	scCities = "#conteudo > div.Prog ul"
)

// City ...
type City struct {
	Name           string `json:"name"`
	FederativeUnit string `json:"ferativeUnit"`
}

// Cities ...
type Cities []City

// GetCities ...
func GetCities() (Cities, error) {
	return New().GetCities()
}

// GetCities ...
func (c *Cinemais) GetCities() (Cities, error) {
	var result Cities
	var err error

	c.collector.OnHTML(scCities, func(e *colly.HTMLElement) {
		// The city list is within a UL element.
		if e.Name == "ul" {
			e.DOM.Find("li").EachWithBreak(func(i int, s *goquery.Selection) bool {
				city, err := parseCity(s)
				if err != nil {
					return false
				}
				result.AppendUniq(city)
				return true
			})
		}
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	c.collector.Visit(c.BaseURL.String())
	return result, err
}

// IsEqual checks whether two city instances are equal or not.
func (c *City) IsEqual(city City) bool {
	return c.Name == city.Name && c.FederativeUnit == city.FederativeUnit
}

// IsValid checks whether the city instance is valid or not.
func (c *City) IsValid() bool {
	return c.Name != "" && IsFederativeUnit(c.FederativeUnit)
}

// AppendUniq ...
func (c *Cities) AppendUniq(city *City) {
	exists := false

	for _, e := range *c {
		if e.IsEqual(*city) {
			exists = true
			break
		}
	}

	if !exists {
		*c = append(*c, *city)
	}
}

// IsFederativeUnit checks if the given string is a federative unit
func IsFederativeUnit(s string) bool {
	if len(s) != 2 {
		return false
	}

	federativeUnits := []string{
		"AC", "AL", "AM", "AP", "BA", "CE",
		"DF", "ES", "GO", "MA", "MG", "MS",
		"MT", "PA", "PB", "PE", "PI", "PR",
		"RJ", "RN", "RO", "RR", "RS", "SC",
		"SE", "SP", "TO",
	}

	for _, f := range federativeUnits {
		if f == s {
			return true
		}
	}

	return false
}

func parseCity(s *goquery.Selection) (*City, error) {
	var result *City
	var err error

	// Get all text from selection ignoring the script
	// containing inside them
	//
	// Retrieve expected texts already trimmed.
	text := util.GetTextTrimmed(s.Contents().Not("script"))
	textParts := strings.Split(text, " - ")

	size := len(textParts)
	if size == 2 {
		result = &City{
			Name:           textParts[0],
			FederativeUnit: textParts[1],
		}
	} else if size == 3 {
		result = &City{
			Name:           textParts[1],
			FederativeUnit: textParts[2],
		}
	}

	if result == nil || !result.IsValid() {
		log.Printf("couldn't get city information from text '%s'\n", text)
		err = ErrNotFound
	}

	return result, err
}
