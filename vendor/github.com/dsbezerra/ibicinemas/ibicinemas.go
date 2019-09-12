package ibicinemas

import (
	"net/url"

	"github.com/gocolly/colly"
)

const (
	baseURL = "http://www.ibicinemas.com.br"
)

const (
	Projection2D = "2D"
	Projection3D = "3D"
)

// Ibicinemas TODO
type Ibicinemas struct {
	BaseURL *url.URL

	collector *colly.Collector
}

// New allocates and initializes a new Ibicinemas instance.
func New() *Ibicinemas {
	i := &Ibicinemas{
		collector: colly.NewCollector(),
	}
	u, _ := url.Parse(baseURL)
	i.BaseURL = u
	return i
}
