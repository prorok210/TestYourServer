package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	DEFAULT_REQ_DELAY              = 100 * time.Millisecond
	MIN_REQ_DELAY                  = 1 * time.Millisecond
	MAX_REQ_DELAY                  = 6000 * time.Millisecond
	DEFAULT_DURATION               = 5 * time.Minute
	MIN_DURATION                   = 1 * time.Minute
	MAX_DURATION                   = 60 * time.Minute
	DEFAULT_COUNT_WORKERS          = 10
	MAX_CCOUNT_WORKERS             = 100
	DEFAULT_REQUEST_CHAN_BUF_SIZE  = 10
	MAX_CHAN_BUF_SIZE              = 100
	DEFAULT_RESPONSE_CHAN_BUF_SIZE = 10
	REPORT_IN_CHAN_SIZE            = 100
	REQUEST_TIMEOUT                = 10 * time.Second
)

type RequestInfo struct {
	Time     time.Duration
	Response *http.Response
	Request  *http.Request
	Err      error
}

type ReqSendingSettings struct {
	Requests            []*http.Request
	Count_Workers       int
	Delay               time.Duration
	Duration            time.Duration
	RequestChanBufSize  int
	ResponseChanBufSize int
}

type CachedRequest struct {
	*http.Request
	cachedBody []byte
}

func StartSendingHttpRequests(outCh chan<- *RequestInfo, reqSettings *ReqSendingSettings, ctx context.Context) []*RequestReport {
	reqSettings = setReqSettings(reqSettings)
	if reqSettings.Requests == nil {
		outCh <- &RequestInfo{Err: errors.New("No requests")}
		return nil
	}

	cachedRequests := make([]*CachedRequest, len(reqSettings.Requests))
	for i, req := range reqSettings.Requests {
		cachedReq := &CachedRequest{Request: req}
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			cachedReq.cachedBody = body
			req.Body.Close()
		}
		cachedRequests[i] = cachedReq
	}

	var reqWg sync.WaitGroup
	reportOutCh := make(chan []*RequestReport, 1)
	reportInCh := make(chan *RequestInfo, REPORT_IN_CHAN_SIZE)

	reqWg.Add(1)
	go func() {
		defer reqWg.Done()
		reportOutCh <- reportPool(ctx, reportInCh)
		close(reportOutCh)
	}()

	var wg sync.WaitGroup
	for i := 0; i < int(reqSettings.Count_Workers); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			customTransport := &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        MAX_CCOUNT_WORKERS,
				MaxIdleConnsPerHost: MAX_CCOUNT_WORKERS,
			}
			cl := http.Client{Transport: customTransport, Timeout: REQUEST_TIMEOUT}

			ticker := time.NewTicker(reqSettings.Delay)
			defer ticker.Stop()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					index := r.Intn(len(cachedRequests))
					cached := cachedRequests[index]

					reqCopy := cached.Request.Clone(ctx)
					if cached.cachedBody != nil {
						reqCopy.Body = io.NopCloser(bytes.NewReader(cached.cachedBody))
					}

					start := time.Now()
					resp, err := cl.Do(reqCopy)

					reqInf := &RequestInfo{
						Time:     time.Since(start),
						Response: resp,
						Request:  cached.Request,
						Err:      err,
					}

					outCh <- reqInf
					reportInCh <- reqInf
				}
			}
		}()
	}

	wg.Wait()
	close(outCh)

	reqWg.Wait()
	close(reportInCh)

	return <-reportOutCh
}
