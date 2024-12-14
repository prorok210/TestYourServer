package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

func CreateAppWindow() {
	a := app.New()
	w := a.NewWindow("Test Your Server")

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("There will be information about the request here...")

	// Test button
	var testButton *widget.Button
	var testIsActiv bool

	// Configurate requests
	var (
		configRequestsButton *widget.Button
		confWindowOpen       bool
		activRequstsRows     []RequestRow
		activRequsts         []http.Request
	)

	scrollContainer := container.NewScroll(entry)

	showRequest := widget.NewCheck("Show Request", nil)
	showTime := widget.NewCheck("Show Time", nil)
	showBody := widget.NewCheck("Show Body (only first 1000 bytes)", nil)
	showHeaders := widget.NewCheck("Show Headers (only first 10 headers)", nil)

	testCtx, testCancel := context.WithCancel(context.Background())
	testButton = widget.NewButton("Start testing", func() {
		if confWindowOpen {
			dialog.ShowInformation("Error", "You can't start testing while the settings window is open", w)
			return
		}
		if testIsActiv {
			testCancel()
			testCtx, testCancel = context.WithCancel(context.Background())
			testButton.SetText("Start testing")
			testIsActiv = false
		} else {
			reqSetting := &core.ReqSendingSettings{
				Requests:            activRequsts,
				Count_Workers:       10,
				Delay:               100 * time.Millisecond,
				Duration:            60 * time.Second,
				RequestChanBufSize:  10,
				ResponseChanBufSize: 10,
			}

			outChan := make(chan *core.RequestInfo, 10)
			go core.StartSendingHttpRequests(outChan, reqSetting, testCtx)

			go func() {
				defer func() {
					testCancel()
					testCtx, testCancel = context.WithCancel(context.Background())
					testButton.SetText("Start testing")
					testIsActiv = false
				}()

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
					entry.SetText(text)
					pendingUpdate = false
				}

				for {
					select {
					case <-testCtx.Done():
						return
					case resp := <-outChan:
						if resp != nil {
							var responseText string
							if resp.Err != nil {
								if resp.Err.Error() == "No requests" {
									dialog.ShowInformation("Error", "No requests", w)
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
	})

	configRequestsButton = widget.NewButton("Configurate requests", func() {
		if testIsActiv {
			dialog.ShowInformation("Error", "You can't configure requests while testing is in progress", w)
			return
		}
		if !confWindowOpen {
			showConfReqWindow(&activRequstsRows, &activRequsts, &confWindowOpen)
			confWindowOpen = true
		}
	})

	w.SetCloseIntercept(func() {
		if confWindowOpen {
			dialog.ShowInformation("Info", "Please close the settings window before exiting.", w)
		} else {
			w.Close()
		}
	})

	optionsContainer := container.NewVBox(showRequest, showTime, showBody, showHeaders)

	w.SetContent(container.NewBorder(container.NewVBox(optionsContainer, testButton, configRequestsButton), nil, nil, nil, scrollContainer))

	w.Resize(fyne.NewSize(1000, 1000))

	w.ShowAndRun()
}
