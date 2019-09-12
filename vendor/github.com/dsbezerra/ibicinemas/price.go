package ibicinemas

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/dsbezerra/ibicinemas/util"
	"github.com/gocolly/colly"
)

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Holiday
	Preview
)

// Day TODO
type Day struct {
	ID   Weekday `json:"id"`
	Name string  `json:"name"`
}

// Price TODO
type Price struct {
	Days       []Day   `json:"days"`
	Projection string  `json:"projection"`
	Full       float32 `json:"full"`
	Half       float32 `json:"half"`
}

// Prices TODO
type Prices []Price

const (
	scPrices = "div.panel-body > table"
)

// GetPrices TODO
func (i *Ibicinemas) GetPrices() (Prices, error) {
	return i.getPrices()
}

// GetPrices TODO
func GetPrices() (Prices, error) {
	return New().getPrices()
}

func (i *Ibicinemas) getPrices() (Prices, error) {
	var result Prices
	var err error
	i.collector.OnHTML(scPrices, func(e *colly.HTMLElement) {
		e.DOM.Find("tr").Each(func(i int, s *goquery.Selection) {
			if i != 0 {
				price, err := parsePrice(s.Find("td"))
				if err == nil {
					result = append(result, *price)
				}
			}
		})
	})
	i.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})
	url := fmt.Sprintf("%s/ibicinemas-tabela-de-precos-4.html", baseURL)
	visitErr := i.collector.Visit(url)
	if visitErr != nil {
		err = visitErr
	}
	return result, err
}

func parsePrice(s *goquery.Selection) (*Price, error) {
	var result Price
	var err error

	parts := strings.Split(strings.TrimSpace(s.Eq(0).Text()), " - ")
	if len(parts) != 2 {
		return nil, errors.New("couldn't find days/projection in text")
	}

	plen := len(parts[0])
	if plen > 0 && plen-2 > 0 {
		projection := parts[0][plen-2:]
		if projection != Projection2D && projection != Projection3D {
			return nil, errors.New("unexpected projection")
		}
		result.Projection = projection
	} else {
		return nil, errors.New("couldn't find projection")
	}

	dlen := len(parts[1])
	if dlen > 0 {
		var days []Day
		for _, d := range strings.Split(parts[1], " ") {
			if len(d) < 3 {
				continue
			}
			d = strings.ToLower(d)
			if strings.Contains(d, "dom") {
				days = append(days, Day{ID: Sunday, Name: "Domingo"})
			} else if strings.Contains(d, "seg") {
				days = append(days, Day{ID: Monday, Name: "Segunda"})
			} else if strings.Contains(d, "ter") {
				days = append(days, Day{ID: Tuesday, Name: "Terça"})
			} else if strings.Contains(d, "qua") {
				days = append(days, Day{ID: Wednesday, Name: "Quarta"})
			} else if strings.Contains(d, "qui") {
				days = append(days, Day{ID: Thursday, Name: "Quinta"})
			} else if strings.Contains(d, "sex") {
				days = append(days, Day{ID: Friday, Name: "Sexta"})
			} else if util.ContainsAny(d, []string{"sáb", "sab"}) {
				days = append(days, Day{ID: Saturday, Name: "Sábado"})
			} else if strings.Contains(d, "fer") {
				days = append(days, Day{ID: Holiday, Name: "Feriado"})
			} else if util.ContainsAny(d, []string{"pré-estreia", "pre-estreia"}) {
				days = append(days, Day{ID: Preview, Name: "Pré-estreia"})
			} else {
				// TODO: Logging
			}
		}
		result.Days = days
	} else {
		return nil, errors.New("couldn't find days")
	}

	full := strings.TrimFunc(s.Eq(1).Text(), func(r rune) bool {
		return unicode.IsSpace(r) || r == 'R' || r == '$'
	})
	full = strings.Replace(full, ",", ".", -1)
	value, err := strconv.ParseFloat(full, 32)
	if err != nil {
		return nil, err
	}
	result.Full = float32(value)
	result.Half = result.Full / 2.0

	return &result, err
}
