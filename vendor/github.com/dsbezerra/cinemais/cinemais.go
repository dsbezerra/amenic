package cinemais

import (
	"errors"
	"net/url"

	"github.com/gocolly/colly"
)

const (
	baseURL = "http://www.cinemais.com.br"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")

	// ErrNoCinemaCode ...
	ErrNoCinemaCode = errors.New("cinema code was not defined")

	// ErrUnexpectedStructure ...
	ErrUnexpectedStructure = errors.New("unexpected DOM structure")
)

// Cinemais ...
type Cinemais struct {
	BaseURL    *url.URL
	CinemaCode string

	collector *colly.Collector
}

// New allocates and initializes a new Cinemais instance.
func New(options ...func(*Cinemais)) *Cinemais {
	c := &Cinemais{
		collector: colly.NewCollector(),
	}
	u, _ := url.Parse(baseURL)
	c.BaseURL = u
	for _, f := range options {
		f(c)
	}
	return c
}

// CinemaCode sets the cinema code
func CinemaCode(cc string) func(*Cinemais) {
	return func(c *Cinemais) {
		c.CinemaCode = cc
	}
}
