package core

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const (
	REP_CHAN_BUF_SIZE = 100
)

type RequestReport struct {
	Url     string
	AvgTime time.Duration
	MinTime time.Duration
	MaxTime time.Duration
	Count   int
	ReqCods map[int]int
	Errors  map[string]int
}

func reportPool(ctx context.Context, in <-chan *RequestInfo) []*RequestReport {
	reqMap := make(map[*http.Request]struct {
		ch  chan *RequestInfo
		rep *RequestReport
	})
	result := make([]*RequestReport, 0)

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return result
		case req := <-in:
			if _, exists := reqMap[req.Request]; !exists {
				repCh := make(chan *RequestInfo, REP_CHAN_BUF_SIZE)
				report := &RequestReport{
					ReqCods: make(map[int]int),
					Errors:  make(map[string]int),
				}

				reqMap[req.Request] = struct {
					ch  chan *RequestInfo
					rep *RequestReport
				}{
					ch:  repCh,
					rep: report,
				}

				result = append(result, report)

				wg.Add(1)
				go func(req *http.Request) {
					calcReport(ctx, repCh, report)
					close(repCh)
					delete(reqMap, req)
					wg.Done()
				}(req.Request)

				repCh <- req

			} else {
				reqMap[req.Request].ch <- req
			}
		}
	}
}

func calcReport(ctx context.Context, in <-chan *RequestInfo, report *RequestReport) {
	var sum time.Duration
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-in:
			if report.Url == "" {
				report.Url = req.Request.URL.String()
			}

			report.Count++
			sum += req.Time

			if report.MinTime == 0 {
				report.MinTime = req.Time
			} else {
				report.MinTime = min(report.MinTime, req.Time)
			}

			report.MaxTime = max(report.MaxTime, req.Time)

			if req.Response != nil {
				report.ReqCods[req.Response.StatusCode]++
			}

			if req.Err != nil {
				report.Errors[req.Err.Error()]++
			}

			report.AvgTime = time.Duration(float64(sum) / float64(report.Count))
		}
	}
}
