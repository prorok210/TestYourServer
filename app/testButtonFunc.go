package app

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"fyne.io/fyne/v2/dialog"
	"github.com/prorok210/TestYourServer/core"
)

const OUT_REQ_CHAN_BUF = 100

func startTesting() {
	delaySlider.Disable()
	durationSlider.Disable()
	delayEntry.Disable()
	durationEntry.Disable()
	workersEntry.Disable()
	workersSlider.Disable()

	testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
}

func endTesting() {
	delaySlider.Enable()
	durationSlider.Enable()
	delayEntry.Enable()
	durationEntry.Enable()
	workersEntry.Enable()
	workersSlider.Enable()
	testCancel()

	testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
	testButton.SetText("Start testing")
	testIsActiv = false
}

func testButtonFunc() {
	if confWindowOpen {
		dialog.ShowInformation("Error", "You can't start testing while the settings window is open", window)
		return
	}

	if testIsActiv {
		endTesting()
	} else {
		startTesting()
		reqSetting := &core.ReqSendingSettings{
			Requests:            activRequsts,
			Count_Workers:       int(workersSlider.Value),
			Delay:               time.Duration(delaySlider.Value) * time.Millisecond,
			Duration:            time.Duration(durationSlider.Value) * time.Second,
			RequestChanBufSize:  10,
			ResponseChanBufSize: 10,
		}

		outChan := make(chan *core.RequestInfo, OUT_REQ_CHAN_BUF)
		// reportChan := make(chan []*core.RequestReport)
		go func() {
			currentReports = core.StartSendingHttpRequests(outChan, reqSetting, testCtx)
		}()

		go func() {
			defer endTesting()

			var lastRequests []string
			maxLines := 20

			updateTicker := time.NewTicker(100 * time.Millisecond)
			defer updateTicker.Stop()

			var pendingUpdate bool

			updateUI := func() {
				if !pendingUpdate {
					return
				}
				text := strings.Join(lastRequests, "\n")
				infoReqsGrid.SetText(text)
				pendingUpdate = false
			}

			for {
				select {
				case <-testCtx.Done():
					return
				case resp, ok := <-outChan:
					if !ok {
						return
					}
					if resp != nil {
						var responseText string
						if resp.Err != nil {
							if resp.Err.Error() == "No requests" {
								dialog.ShowInformation("Error", "No requests", window)
								return
							}
							responseText += fmt.Sprintf("Error: %v\n", resp.Err)
							continue
						}

						if showRequest.Checked {
							responseText += fmt.Sprintf("Request: %v %v\n", resp.Request.Method, resp.Request.URL)
						}
						if showTime.Checked {
							responseText += fmt.Sprintf("Time: %v\n", resp.Time)
						}
						responseText += fmt.Sprintf("Status: %s\n", resp.Response.Status)

						if showBody.Checked {
							if resp.Response.Body != nil {

								body, err := io.ReadAll(resp.Response.Body)
								if err == nil && len(body) > 0 {
									if len(body) > 1000 {
										responseText += fmt.Sprintf("ResponseBody: %s\n", string(body[:1000]))
									} else {
										responseText += fmt.Sprintf("ResponseBody: %s\n", string(body))
									}
								} else {
									responseText += fmt.Sprintf("ResponseBody: Error reading body\n")
								}
							} else {
								responseText += fmt.Sprintf("ResponseBody: nil\n")
							}
						}
						if showHeaders.Checked {
							if resp.Response.Header != nil {
								responseText += "Headers:\n"
								counter := 0
								for k, v := range resp.Response.Header {
									responseText += fmt.Sprintf("%s: %s\n", k, v)
									counter++
									if counter >= 10 {
										responseText += fmt.Sprintf("...")
										break
									}
								}
							} else {
								responseText += "Headers: nil\n"
							}
						}

						lastRequests = append(lastRequests, responseText)
						if len(lastRequests) > maxLines {
							lastRequests = lastRequests[1:]
						}
						if resp.Response != nil {
							if resp.Response.Body != nil {
								resp.Response.Body.Close()
							}
						}
						pendingUpdate = true
					}
				case <-updateTicker.C:
					updateUI()
				}
			}
		}()
		testButton.SetText("Stop testing")
		testIsActiv = true
	}
}
