package uriutil

import (
	"log"
	"net/url"
)

// ResolveURL resolves a relative url to absolute
func ResolveURL(base, relative string) string {
	// Converts relative string url to URL type
	u, err := url.Parse(relative)
	if err != nil {
		log.Fatal(err)
	}

	// Convert base string url to URL type
	bu, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}

	// Resolves the relative to absolute url as string
	return bu.ResolveReference(u).String()
}
