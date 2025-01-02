package core

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

func MustParseURL(rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New("invalid URL")
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("URL must have a valid scheme and host")
	}
	return parsedURL, nil
}

func ValidateURL(rawURL string) error {
	_, err := MustParseURL(rawURL)
	return err
}

func setReqSettings(reqSettings *RequestsConfig) *RequestsConfig {
	if reqSettings == nil {
		return &RequestsConfig{
			Requests:            nil,
			Count_Workers:       DEFAULT_COUNT_WORKERS,
			Delay:               DEFAULT_REQ_DELAY,
			Duration:            DEFAULT_DURATION,
			RequestChanBufSize:  DEFAULT_REQUEST_CHAN_BUF_SIZE,
			ResponseChanBufSize: DEFAULT_RESPONSE_CHAN_BUF_SIZE,
		}
	}

	if reqSettings.Count_Workers == 0 || reqSettings.Count_Workers > 100 {
		reqSettings.Count_Workers = DEFAULT_COUNT_WORKERS
	}
	if reqSettings.Delay == 0 || reqSettings.Delay > 60*time.Second {
		reqSettings.Delay = DEFAULT_REQ_DELAY
	}
	if reqSettings.Duration == 0 || reqSettings.Duration > 60*time.Minute {
		reqSettings.Duration = DEFAULT_DURATION
	}
	if reqSettings.RequestChanBufSize == 0 || reqSettings.RequestChanBufSize > 100 {
		reqSettings.RequestChanBufSize = DEFAULT_REQUEST_CHAN_BUF_SIZE
	}
	if reqSettings.ResponseChanBufSize == 0 || reqSettings.ResponseChanBufSize > 100 {
		reqSettings.ResponseChanBufSize = DEFAULT_RESPONSE_CHAN_BUF_SIZE
	}
	return reqSettings
}

func TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength] + "..."
}

func WrapText(text string, limit int) string {
	var result strings.Builder
	for len(text) > limit {
		result.WriteString(text[:limit])
		result.WriteString("\n")
		text = text[limit:]
	}
	result.WriteString(text)
	return result.String()
}
