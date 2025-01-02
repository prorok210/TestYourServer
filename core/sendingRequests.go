package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DEFAULT_PROTO                  = HTTP
	DEFAULT_REQ_DELAY              = 100 * time.Millisecond
	MIN_REQ_DELAY                  = 1 * time.Millisecond
	MAX_REQ_DELAY                  = 6000 * time.Millisecond
	DEFAULT_DURATION               = 5 * time.Minute
	MIN_DURATION                   = 1 * time.Minute
	MAX_DURATION                   = 60 * time.Minute
	DEFAULT_COUNT_WORKERS          = 10
	MAX_COUNT_WORKERS              = 100
	DEFAULT_REQUEST_CHAN_BUF_SIZE  = 10
	MAX_CHAN_BUF_SIZE              = 100
	DEFAULT_RESPONSE_CHAN_BUF_SIZE = 10
	REPORT_IN_CHAN_SIZE            = 100
	REQUEST_TIMEOUT                = 10 * time.Second
)

func StartSendingRequests(outCh chan<- *RequestInfo, reqsConfig *RequestsConfig, testCtx context.Context) []*RequestReport {
	reqsConfig = setReqSettings(reqsConfig)
	if reqsConfig.Requests == nil {
		outCh <- &RequestInfo{Err: errors.New("No requests")}
		return nil
	}

	var reportWg sync.WaitGroup
	var sendingReqsWg sync.WaitGroup

	reportOutCh := make(chan []*RequestReport, 1)
	reportInCh := make(chan *RequestInfo, REPORT_IN_CHAN_SIZE)

	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		reportOutCh <- reportPool(reportInCh)
		close(reportOutCh)
	}()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < int(reqsConfig.Count_Workers); i++ {
		sendingReqsWg.Add(1)

		handleHTTP := func() {
			defer sendingReqsWg.Done()
			customTransport := &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: reqsConfig.Secure},
				MaxIdleConns:        MAX_COUNT_WORKERS,
				MaxIdleConnsPerHost: MAX_COUNT_WORKERS,
			}
			cl := http.Client{Transport: customTransport, Timeout: REQUEST_TIMEOUT}

			ticker := time.NewTicker(reqsConfig.Delay)
			defer ticker.Stop()

			for {
				select {
				case <-testCtx.Done():
					return
				case <-ticker.C:
					index := r.Intn(len(reqsConfig.Requests))
					req, ok := reqsConfig.Requests[index].(*HTTPRequest)
					if !ok {
						outCh <- &RequestInfo{Err: errors.New("Unsupported request type")}
						return
					}
					cached := reqsConfig.Requests[index].GetBody()

					reqCopy := req.Clone(testCtx)
					if cached != nil {
						reqCopy.Body = io.NopCloser(bytes.NewReader(cached))
					}

					start := time.Now()
					resp, err := cl.Do(reqCopy)
					if err != nil && strings.Contains(err.Error(), "context canceled") {
						return
					}
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					reqInf := &RequestInfo{
						Time:     time.Since(start),
						Response: &Response{Status: resp.StatusCode, Body: body, Headers: resp.Header},
						Request:  req,
						Err:      err,
					}

					select {
					case outCh <- reqInf:
					default:
						fmt.Println("outCh is full, dropping request")
					}

					select {
					case reportInCh <- reqInf:
					default:
						fmt.Println("reportInCh is full, dropping request")
					}

				}
			}
		}

		handleWebSocket := func() {
			defer sendingReqsWg.Done()

			index := r.Intn(len(reqsConfig.Requests))
			req, ok := reqsConfig.Requests[index].(*WSRequest)
			if !ok {
				outCh <- &RequestInfo{Err: errors.New("Unsupported request type")}
				return
			}

			dialer := websocket.Dialer{
				TLSClientConfig:  &tls.Config{InsecureSkipVerify: reqsConfig.Secure},
				HandshakeTimeout: REQUEST_TIMEOUT,
			}

			conn, _, err := dialer.Dial(reqsConfig.Requests[index].GetURI(), reqsConfig.Requests[index].GetHeaders())
			defer conn.Close()
			if err != nil {
				outCh <- &RequestInfo{Err: err}
				return
			}

			ticker := time.NewTicker(reqsConfig.Delay)
			defer ticker.Stop()

			for {
				select {
				case <-testCtx.Done():
					return
				case <-ticker.C:
					start := time.Now()

					err := conn.WriteMessage(websocket.TextMessage, req.GetBody())
					if err != nil {
						outCh <- &RequestInfo{Err: err}
						return
					}

					msgType, msg, err := conn.ReadMessage()
					if err != nil {
						outCh <- &RequestInfo{Err: err}
						return
					}

					duration := time.Since(start)

					reqInf := &RequestInfo{
						Time:     duration,
						Response: &Response{Status: msgType, Body: msg},
						Request:  req,
						Err:      err,
					}

					select {
					case outCh <- reqInf:
					default:
						fmt.Println("outCh is full, dropping request")
					}

					select {
					case reportInCh <- reqInf:
					default:
						fmt.Println("reportInCh is full, dropping request")
					}
				}
			}
		}

		switch reqsConfig.Protocol {
		case HTTP:
			go handleHTTP()
		case WS:
			go handleWebSocket()
		default:
			outCh <- &RequestInfo{Err: errors.New("Unsupported protocol")}
			return nil
		}

	}

	sendingReqsWg.Wait()
	close(reportInCh)
	close(outCh)
	reportWg.Wait()

	return <-reportOutCh
}
