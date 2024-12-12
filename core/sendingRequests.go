package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type RequestInfo struct {
	Time     time.Duration
	Response *http.Response
}

const COUNT_WORKERS = 10

func StartSendingHttpRequests(outCh chan<- *RequestInfo, ctx context.Context) {
	reqChanMap := make(map[chan *http.Request]struct{})

	for i := 0; i < COUNT_WORKERS; i++ {
		rqCh := make(chan *http.Request, 5)
		reqChanMap[rqCh] = struct{}{}

		customTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cl := http.Client{Transport: customTransport}

		go sendReqLoop(cl, rqCh, outCh, ctx)
	}

	requests := []*http.Request{
		{
			Method: "GET",
			URL:    mustParseURL("https://193.233.114.35:8446/image/get"),
		},
		{
			Method: "POST",
			URL:    mustParseURL("https://193.233.114.35:8446/user/create"),
		},
		{
			Method: "GET",
			URL:    mustParseURL("https://193.233.114.35:8446/user/me"),
		},
		{
			Method: "GET",
			URL:    mustParseURL("https://193.233.114.35:8446/user/get/1"),
		},
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for reqCh := range reqChanMap {
					index := r.Intn(len(requests))
					reqCh <- requests[index]
					time.Sleep(time.Millisecond * 10)
				}
			}

		}
	}()
}

func sendReqLoop(cl http.Client, req <-chan *http.Request, out chan<- *RequestInfo, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rq := <-req:
			start := time.Now()
			r, err := cl.Do(rq)
			if err != nil {
				fmt.Println(err)
			}
			reqInfo := &RequestInfo{
				time.Since(start),
				r,
			}
			out <- reqInfo
		}
	}
}
