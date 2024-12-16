package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	DEFAULT_DELAY                  = 100 * time.Millisecond
	DEFAULT_DURATION               = 60 * time.Second
	DEFAULT_COUNT_WORKERS          = 10
	DEFAULT_REQUEST_CHAN_BUF_SIZE  = 10
	DEFAULT_RESPONSE_CHAN_BUF_SIZE = 10
)

type RequestInfo struct {
	Time     time.Duration
	Response *http.Response
	Request  *http.Request
	Err      error
}

type ReqSendingSettings struct {
	Requests            []http.Request
	Count_Workers       uint
	Delay               time.Duration
	Duration            time.Duration
	RequestChanBufSize  uint
	ResponseChanBufSize uint
}

func StartSendingHttpRequests(outCh chan<- *RequestInfo, reqSettings *ReqSendingSettings, ctx context.Context) {
	reqSettings = setReqSettings(reqSettings)
	if reqSettings.Requests == nil {
		outCh <- &RequestInfo{Err: errors.New("No requests")}
		return
	}

	reqChanMap := make(map[chan *http.Request]struct{})

	for i := 0; i < int(reqSettings.Count_Workers); i++ {
		rqCh := make(chan *http.Request, reqSettings.RequestChanBufSize)
		reqChanMap[rqCh] = struct{}{}

		customTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cl := http.Client{Transport: customTransport}

		go sendReqLoop(cl, rqCh, outCh, ctx)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for reqCh := range reqChanMap {
					index := r.Intn(len(reqSettings.Requests))
					reqCh <- &reqSettings.Requests[index]
					reqSettings.Requests[index].Body.Close()
					time.Sleep(reqSettings.Delay)
				}
			}

		}
	}()
}

func sendReqLoop(cl http.Client, reqCh <-chan *http.Request, outCh chan<- *RequestInfo, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rq := <-reqCh:
			reqCopy := new(http.Request)
			*reqCopy = *rq

			if rq.Body != nil {
				bodyBytes, err := io.ReadAll(rq.Body)
				if err != nil {
					outCh <- &RequestInfo{Err: err}
					continue
				}
				rq.Body.Close()

				rq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				reqCopy.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			start := time.Now()
			resp, err := cl.Do(reqCopy)

			outCh <- &RequestInfo{
				time.Since(start),
				resp,
				rq,
				err,
			}
		}
	}
}
