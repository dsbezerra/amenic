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
	// Selector of the container that contains promotion info.
	scPromotion = "#promoContainer"

	// Selector of the container that contains promotion list.
	scPromotions = "#promoContainer > div.promoList"

	// Selector of the container of promotion page.
	scPromotionPage = "#promoContainer > div.PromoPage"

	// Selector of the promotion item in the list.
	sPromotionItem = "#promoContainer > div.promoList > div.promoInfo"

	// Selector of the promotion page link.
	sPromotionLink = "a"
)

// Promotion ...
type Promotion struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Limited     bool   `json:"limited"`
}

// Promotions ...
type Promotions []Promotion

// GetPromotions ...
func GetPromotions() (Promotions, error) {
	return New().GetPromotions()
}

// GetPromotions ...
func (c *Cinemais) GetPromotions() (Promotions, error) {
	var result Promotions
	var err error

	c.collector.OnHTML(sPromotionItem, func(e *colly.HTMLElement) {
		href, exists := e.DOM.Find("a").Attr("href")
		if exists {
			url := e.Request.URL.String()
			if !strings.Contains(href, url) {
				url = fmt.Sprintf("%s%s", url, href)
			}

			c.collector.OnHTML(scPromotion, func(e *colly.HTMLElement) {
				p, err := parsePromotion(e)
				if err != nil {
					return
				}
				result.AppendUniq(p)
			})

			c.collector.Visit(url)
		}
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/promocoes", c.BaseURL.String())
	c.collector.Visit(url)

	return result, err
}

// IsEqual checks whether two promotion instances are equal or not.
func (p *Promotion) IsEqual(promotion Promotion) bool {
	return p.ID == promotion.ID && p.Name == promotion.Name
}

// IsValid checks whether the promotion instance is valid or not.
func (p *Promotion) IsValid() bool {
	return p.ID != 0 && p.Name != "" && p.Description != ""
}

// AppendUniq ...
func (p *Promotions) AppendUniq(promotion *Promotion) {
	exists := false

	for _, e := range *p {
		if e.IsEqual(*promotion) {
			exists = true
			break
		}
	}

	if !exists {
		*p = append(*p, *promotion)
	}
}

func parsePromotion(e *colly.HTMLElement) (*Promotion, error) {
	var result *Promotion
	var err error

	ID, err := strconv.Atoi(e.Request.URL.Query().Get("cp"))
	if err != nil {
		return nil, err
	}

	result = &Promotion{ID: ID}

	s := e.DOM
	name := util.GetTextTrimmed(s.Find("h1"))
	parts := strings.Split(name, "\"")
	if len(parts) == 3 {
		name = parts[1]
	} else {
		log.Printf("couldn't find name between quotes in text %s\n", name)
	}

	result.Name = name

	image := s.Find("img").AttrOr("src", "")
	if image == "" {
		log.Println("couldn't find promotion image")
	}

	result.Image = image

	s = s.Find("p").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.AttrOr("style", "") == ""
	})

	size := s.Length()

	var description string
	s.Each(func(i int, s *goquery.Selection) {
		t := util.GetTextTrimmed(s.Contents().Not("br"))
		if strings.Contains(t, "é válida por tempo indeterminado") {
			result.Limited = true
		}

		description += t
		if i < size-1 {
			description += "\n"
		}
	})
	result.Description = strings.TrimSpace(description)

	if result == nil || !result.IsValid() {
		log.Printf("couldn't get promotion information from text: %s\n", s.Text())
		err = ErrNotFound
	}

	return result, err
}
