package util

import "net/url"

// ResolveRelativeURL ...
func ResolveRelativeURL(rawbase string, rawrelative string) (string, error) {
	baseURL, err := url.Parse(rawbase)
	if err != nil {
		return "", err
	}
	relativeURL, err := url.Parse(rawrelative)
	if err != nil {
		return "", err
	}
	return baseURL.ResolveReference(relativeURL).String(), err
}
