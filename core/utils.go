package core

import (
	"fmt"
	"net/url"
)

func mustParseURL(rawURL string) *url.URL {
	url, err := url.ParseRequestURI(rawURL)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL: %v", err))
	}
	return url
}
