package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/dialog"
	"github.com/prorok210/TestYourServer/core"
)

const (
	OUT_REQ_CHAN_BUF       = 100
	UPDATE_TEXT_GRID_DALEY = 200 * time.Millisecond
	MAX_LINES              = 20
	MAX_BODY_LEN           = 1000
	MAX_ENTRY_LEN          = 2000
	MAX_HEADERS            = 10
)

func startTesting() {
	delaySlider.Disable()
	durationSlider.Disable()
	delayEntry.Disable()
	durationEntry.Disable()
	workersEntry.Disable()
	workersSlider.Disable()
	reportButton.Disable()
	configRequestsButton.Disable()
	protocolButton.Disable()

	testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
	displayCtx, displayCtxCancel = context.WithCancel(context.Background())
	countReqs.Store(0)
	countFailedReqs.Store(0)
	go startTimer(time.Duration(durationSlider.Value) * time.Minute)
}

func endTesting() {
	testCancel()
	<-displayCtx.Done()

	testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
	displayCtx, displayCtxCancel = context.WithCancel(context.Background())
	testButton.SetText("Start testing")
	testIsActiv = false

	delaySlider.Enable()
	durationSlider.Enable()
	delayEntry.Enable()
	durationEntry.Enable()
	workersEntry.Enable()
	workersSlider.Enable()
	reportButton.Enable()
	configRequestsButton.Enable()
	protocolButton.Enable()
}

func testButtonFunc() {
	if confWindowOpen {
		dialog.ShowInformation("Error", "Can't start testing while the settings window is open", window)
		return
	}

	if protocolWindowOpen {
		dialog.ShowInformation("Error", "Can't start testing while the protocol window is open", window)
		return
	}

	if len(activRequsts) == 0 {
		dialog.ShowInformation("Error", "Configure requests before starting the test", window)
		return
	}

	if testIsActiv {
		testCancel()
	} else {
		startTesting()
		reqSetting := &core.RequestsConfig{
			Requests:            activRequsts,
			Count_Workers:       int(workersSlider.Value),
			Delay:               time.Duration(delaySlider.Value) * time.Millisecond,
			Duration:            time.Duration(durationSlider.Value) * time.Second,
			RequestChanBufSize:  100,
			ResponseChanBufSize: 100,
		}

		outChan := make(chan *core.RequestInfo, OUT_REQ_CHAN_BUF)

		go func() {
			currentReports = core.StartSendingHttpRequests(outChan, reqSetting, testCtx)
			displayCtxCancel()
		}()

		ringBuffer := make([]string, MAX_LINES)
		currentIndex := 0

		go func() {
			defer endTesting()

			updateTicker := time.NewTicker(UPDATE_TEXT_GRID_DALEY)
			defer updateTicker.Stop()

			var mu sync.Mutex
			var pendingUpdate bool
			var batchText strings.Builder

			updateUI := func() {
				if !pendingUpdate {
					return
				}
				mu.Lock()
				var displayText strings.Builder
				for i := 0; i < MAX_LINES; i++ {
					idx := (currentIndex - i - 1 + MAX_LINES) % MAX_LINES
					if ringBuffer[idx] != "" {
						displayText.WriteString(ringBuffer[idx])
						displayText.WriteString("\n")
					}
				}
				text := displayText.String()
				mu.Unlock()

				if text != "" {
					infoReqsGrid.SetText(text)
				}
				pendingUpdate = false
			}

			for {
				select {
				case <-displayCtx.Done():
					return
				case resp, ok := <-outChan:
					if !ok {
						return
					}
					batchText.Reset()

					countReqs.Add(1)

					if resp.Err != nil {
						if resp.Err.Error() == "No requests" {
							dialog.ShowInformation("Error", "No requests", window)
							return
						}
						countFailedReqs.Add(1)
						batchText.WriteString(fmt.Sprintf("Error: %v\n", core.TruncateString(resp.Err.Error(), MAX_ROW_LEN)))
					}

					if showRequest.Checked {
						fmt.Fprintf(&batchText, "Request: %v %v\n", resp.Request.GetMethod(), core.TruncateString(resp.Request.GetURI(), MAX_ROW_LEN))
					}

					if showTime.Checked {
						fmt.Fprintf(&batchText, "Time: %v\n", resp.Time)
					}

					if resp.Response != nil {
						fmt.Fprintf(&batchText, "Status: %s\n", resp.Response.Status)
						if showBody.Checked && resp.Response.Body != nil {
							body, err := io.ReadAll(resp.Response.Body)
							resp.Response.Body.Close()

							if err == nil && len(body) > 0 {
								bodyStr := core.WrapText(string(body), MAX_ROW_LEN)
								if len(bodyStr) > MAX_BODY_LEN {
									bodyStr = bodyStr[:MAX_BODY_LEN] + "..."
								}
								fmt.Fprintf(&batchText, "ResponseBody: %s\n", bodyStr)
							}

						}

						if showHeaders.Checked && resp.Response.Header != nil {
							batchText.WriteString("Headers:\n")
							headerCount := 0
							for k, v := range resp.Response.Header {
								if headerCount >= MAX_HEADERS {
									batchText.WriteString("...\n")
									break
								}
								fmt.Fprintf(&batchText, "%s: %s\n", k, v)
								headerCount++
							}
						}
					}

					entryText := batchText.String()
					if len(entryText) > MAX_ENTRY_LEN {
						entryText = entryText[:MAX_ENTRY_LEN] + "..."
					}

					mu.Lock()
					ringBuffer[currentIndex] = entryText
					currentIndex = (currentIndex + 1) % MAX_LINES
					mu.Unlock()

					pendingUpdate = true

				case <-updateTicker.C:
					updateUI()
				}
			}
		}()
		testButton.SetText("Stop testing")
		testIsActiv = true
	}
}

func startTimer(maxDuration time.Duration) {
	elapsed := time.Duration(0)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var remainingMinutes, remainingSeconds, elapsedMinutes, elapsedSeconds int

	for {
		select {
		case <-ticker.C:
			elapsed += time.Second
			remaining := maxDuration - elapsed
			if remaining <= 0 {
				return
			}
			remainingMinutes = int(remaining.Minutes())
			remainingSeconds = int(remaining.Seconds()) % 60
			elapsedMinutes = int(elapsed.Minutes())
			elapsedSeconds = int(elapsed.Seconds()) % 60
			StatsLabel.SetText(fmt.Sprintf("Time left: %02d:%02d\nTime elapsed: %02d:%02d\nRequests sent: %d\nRequests failed: %d",
				remainingMinutes, remainingSeconds, elapsedMinutes, elapsedSeconds, countReqs.Load(), countFailedReqs.Load()))
		case <-displayCtx.Done():
			StatsLabel.SetText(fmt.Sprintf("Time left: %02d:%02d\nTime elapsed: %02d:%02d\nRequests sent: %d\nRequests failed: %d",
				remainingMinutes, remainingSeconds, elapsedMinutes, elapsedSeconds, countReqs.Load(), countFailedReqs.Load()))
			return
		}
	}
}
