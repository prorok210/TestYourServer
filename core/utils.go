package core

import (
	"net/url"
	"time"
)

func MustParseURL(rawURL string) *url.URL {
	url, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil
	}
	return url
}

func setReqSettings(reqSettings *ReqSendingSettings) *ReqSendingSettings {
	if reqSettings == nil {
		return &ReqSendingSettings{
			Requests:            nil,
			Count_Workers:       DEFAULT_COUNT_WORKERS,
			Delay:               DEFAULT_DELAY,
			Duration:            DEFAULT_DURATION,
			RequestChanBufSize:  DEFAULT_REQUEST_CHAN_BUF_SIZE,
			ResponseChanBufSize: DEFAULT_RESPONSE_CHAN_BUF_SIZE,
		}
	}

	if reqSettings.Count_Workers == 0 || reqSettings.Count_Workers > 100 {
		reqSettings.Count_Workers = DEFAULT_COUNT_WORKERS
	}
	if reqSettings.Delay == 0 || reqSettings.Delay > 60*time.Second {
		reqSettings.Delay = DEFAULT_DELAY
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
