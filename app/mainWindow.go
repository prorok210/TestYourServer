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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

func CreateAppWindow() {
	a := app.New()
	w := a.NewWindow("Test Your Server")

	infoReqsGrid := widget.NewTextGrid()
	infoReqsGrid.SetText("There will be information about the requests here...")

	// Test button
	var (
		testButton  *widget.Button
		testIsActiv bool
	)

	// Configurate requests
	var (
		configRequestsButton *widget.Button
		confWindowOpen       bool
		activRequstsRows     []RequestRow
		activRequsts         []http.Request
	)

	// Protocol selection
	var (
		protocolButton *widget.Button = widget.NewButton("Select protocol", func() {})
	)

	// Report button
	var (
		reportButton *widget.Button = widget.NewButton("Show report", func() {})
	)

	// Sliders for delay and duration
	delaySlider := widget.NewSlider(1, 6000)
	delaySlider.Step = 10
	delaySlider.SetValue(200)

	durationSlider := widget.NewSlider(1, 60)
	durationSlider.Step = 0.5
	durationSlider.SetValue(5)

	// Entry for delay and duration
	delayEntry := widget.NewEntry()
	delayEntry.SetText("200 ms")
	delayEntry.Resize(fyne.NewSize(100, 1000))
	durationEntry := widget.NewEntry()
	durationEntry.SetText("5 min")
	durationEntry.Resize(fyne.NewSize(100, 1000))

	// Options for showing request
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
					infoReqsGrid.SetText(text)
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

	// horizontal container for delay and duration
	delayContainer := container.NewHBox(
		widget.NewLabel("Request delay"),
		container.NewGridWrap(fyne.NewSize(300, 40), delaySlider),
		container.NewGridWrap(fyne.NewSize(70, 40), delayEntry),
	)
	durationContainer := container.NewHBox(
		widget.NewLabel("Test duration  "),
		container.NewGridWrap(fyne.NewSize(300, 40), durationSlider),
		container.NewGridWrap(fyne.NewSize(70, 40), durationEntry),
	)

	// Left panel of options
	optionsContainer := container.NewVBox(
		container.NewVBox(
			widget.NewLabel("Testing options"),
			showRequest,
			showBody,
			showHeaders,
			showTime,
		),

		layout.NewSpacer(),
		container.NewVBox(
			delayContainer,
			durationContainer,
		),
		layout.NewSpacer(),
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(35, 40), container.New(layout.NewHBoxLayout())),
			container.NewGridWrap(fyne.NewSize(150, 40), protocolButton),
			container.NewGridWrap(fyne.NewSize(100, 40), container.New(layout.NewHBoxLayout())),
			container.NewGridWrap(fyne.NewSize(150, 40), configRequestsButton),
		),
		layout.NewSpacer(),
	)

	// Scroll container
	scrollContainer := container.NewScroll(infoReqsGrid)

	top := container.NewHSplit(optionsContainer, scrollContainer)
	top.SetOffset(0.3)

	content := container.NewVSplit(
		top,
		container.NewAdaptiveGrid(5,
			layout.NewSpacer(),
			container.NewGridWrap(fyne.NewSize(150, 40), reportButton),
			layout.NewSpacer(),
			container.NewGridWrap(fyne.NewSize(150, 40), testButton),
			layout.NewSpacer(),
		),
	)
	content.SetOffset(0.95)

	w.SetContent(content)

	w.Resize(fyne.NewSize(1200, 600))

	w.ShowAndRun()
}
