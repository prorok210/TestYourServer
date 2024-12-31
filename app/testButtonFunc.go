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
	// protocolButton.Disable()

	testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
	displayCtx, displayCtxCancel = context.WithCancel(context.Background())
}

func endTesting() {
	fmt.Println("End testing")
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
	// protocolButton.Enable()
}

func testButtonFunc() {
	if confWindowOpen {
		dialog.ShowInformation("Error", "Can't start testing while the settings window is open", window)
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
		reqSetting := &core.ReqSendingSettings{
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

					if resp.Err != nil {
						if resp.Err.Error() == "No requests" {
							dialog.ShowInformation("Error", "No requests", window)
							return
						}
						batchText.WriteString(fmt.Sprintf("Error: %v\n", resp.Err))
						continue
					}

					if showRequest.Checked {
						fmt.Fprintf(&batchText, "Request: %v %v\n", resp.Request.Method, resp.Request.URL)
					}

					if showTime.Checked {
						fmt.Fprintf(&batchText, "Time: %v\n", resp.Time)
					}

					if resp.Response != nil {
						fmt.Fprintf(&batchText, "Status: %s\n", resp.Response.Status)
						if showBody.Checked && resp.Response.Body != nil {
							body, err := io.ReadAll(resp.Response.Body)
							if err == nil && len(body) > 0 {
								bodyStr := string(body)
								if len(bodyStr) > 1000 {
									bodyStr = bodyStr[:1000] + "..."
								}
								fmt.Fprintf(&batchText, "ResponseBody: %s\n", bodyStr)
							}
							resp.Response.Body.Close()
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
