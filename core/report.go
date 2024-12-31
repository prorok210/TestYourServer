package core

import (
	"net/http"
	"sync"
	"time"
)

const (
	REP_CHAN_BUF_SIZE = 100
)

type RequestReport struct {
	Url      string
	AvgTime  time.Duration
	MinTime  time.Duration
	MaxTime  time.Duration
	Count    int
	ReqCods  map[int]int
	Errors   map[string]int
	TestTime time.Time
}

func reportPool(in <-chan *RequestInfo) []*RequestReport {
	reqMap := make(map[*http.Request]struct {
		ch  chan *RequestInfo
		rep *RequestReport
	})
	result := make([]*RequestReport, 0)

	var calcRepWg sync.WaitGroup

	for {
		req, ok := <-in
		if !ok {
			for _, v := range reqMap {
				close(v.ch)
			}
			calcRepWg.Wait()

			return result
		}

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

			calcRepWg.Add(1)
			go func() {
				calcReportLoop(repCh, report)

				defer func(req *http.Request) {
					calcRepWg.Done()
					delete(reqMap, req)
				}(req.Request)
			}()

			repCh <- req
		} else {
			reqMap[req.Request].ch <- req
		}
	}
}

func calcReportLoop(in <-chan *RequestInfo, report *RequestReport) {
	var sum time.Duration
	for {
		req, ok := <-in
		if !ok {
			return
		}
		calcReport(&sum, req, report)
	}
}

func calcReport(sum *time.Duration, req *RequestInfo, report *RequestReport) {
	if report.Url == "" {
		report.Url = req.Request.URL.String()
	}

	report.Count++
	*sum += req.Time

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

	report.AvgTime = time.Duration(float64(*sum) / float64(report.Count))
}
