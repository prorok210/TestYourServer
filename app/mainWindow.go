package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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

const (
	MAX_LINES          int     = 20
	REQ_DELAY_STEP     float64 = 1
	TEST_DURATION_STEP float64 = 0.5
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
		activRequstsRows     []*RequestRow
		activRequsts         []*http.Request
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
	delaySlider := widget.NewSlider(float64(core.MIN_REQ_DELAY.Milliseconds()), float64(core.MAX_REQ_DELAY.Milliseconds()))
	delaySlider.Step = REQ_DELAY_STEP
	delaySlider.SetValue(float64(core.DEFAULT_REQ_DELAY.Milliseconds()))
	delayValStr := fmt.Sprintf("%v ms", core.DEFAULT_REQ_DELAY.Milliseconds())

	durationSlider := widget.NewSlider(float64(core.MIN_DURATION.Minutes()), float64(core.MAX_DURATION.Minutes()))
	durationSlider.Step = TEST_DURATION_STEP
	durationSlider.SetValue(float64(core.DEFAULT_DURATION.Minutes()))
	durationValStr := fmt.Sprintf("%v min", core.DEFAULT_DURATION.Minutes())

	// Entry for delay and duration
	delayEntry := widget.NewEntry()
	delayEntry.SetText(delayValStr)
	delayEntry.Resize(fyne.NewSize(100, 1000))

	durationEntry := widget.NewEntry()
	durationEntry.SetText(durationValStr)
	durationEntry.Resize(fyne.NewSize(100, 1000))

	// OnChanged for sliders
	delaySlider.OnChanged = func(f float64) {
		delayValStr = fmt.Sprintf("%v ms", f)
		delayEntry.SetText(delayValStr)
	}
	durationSlider.OnChanged = func(f float64) {
		durationValStr = fmt.Sprintf("%v min", f)
		durationEntry.SetText(durationValStr)
	}

	// OnChanged for entries
	delayEntry.OnChanged = func(s string) {
		s = strings.TrimSuffix(s, " ms")
		matched, _ := regexp.MatchString(`^[0-9]+$`, s)

		if matched {
			val, _ := strconv.ParseFloat(s, 64)
			if val > float64(core.MAX_REQ_DELAY.Milliseconds()) {
				val = float64(core.MAX_REQ_DELAY.Milliseconds())
			}
			if val < float64(core.MIN_REQ_DELAY.Milliseconds()) {
				val = float64(core.MIN_REQ_DELAY.Milliseconds())
			}
			delaySlider.SetValue(val)
			delayValStr = fmt.Sprintf("%v ms", val)
			delayEntry.SetText(delayValStr)
		} else {
			delayValStr = fmt.Sprintf("%v ms", core.MIN_REQ_DELAY.Milliseconds())
			delayEntry.SetText(delayValStr)
		}
	}

	durationEntry.OnChanged = func(s string) {
		s = strings.TrimSuffix(s, " min")
		matched, _ := regexp.MatchString(`^[0-9]+(\.[0-9]+)?`, s)

		if matched {
			val, _ := strconv.ParseFloat(s, 64)
			if val < float64(core.MIN_DURATION.Minutes()) {
				val = float64(core.MIN_DURATION.Minutes())
			}
			if val > float64(core.MAX_DURATION.Minutes()) {
				val = float64(core.MAX_DURATION.Minutes())
			}
			durationSlider.SetValue(val)
			durationValStr = fmt.Sprintf("%v min", val)
			durationEntry.SetText(durationValStr)
		} else {
			durationValStr = fmt.Sprintf("%v min", 1)
			durationEntry.SetText(durationValStr)
		}
	}

	// Options for showing request
	showRequest := widget.NewCheck("Show request", nil)
	showTime := widget.NewCheck("Show response Time", nil)
	showBody := widget.NewCheck("Show response Body (only first 1000 bytes)", nil)
	showHeaders := widget.NewCheck("Show response Headers (only first 10 headers)", nil)

	testCtx, testCancel := context.Background(), func() {}

	testButton = widget.NewButton("Start testing", func() {
		if confWindowOpen {
			dialog.ShowInformation("Error", "You can't start testing while the settings window is open", w)
			return
		}
		startTestint := func() {
			delaySlider.Disable()
			durationSlider.Disable()
			delayEntry.Disable()
			durationEntry.Disable()

			testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
		}
		endTesting := func() {
			delaySlider.Enable()
			durationSlider.Enable()
			delayEntry.Enable()
			durationEntry.Enable()

			testCancel()
			testCtx, testCancel = context.WithTimeout(context.Background(), time.Duration(durationSlider.Value)*time.Minute)
			testButton.SetText("Start testing")
			testIsActiv = false
		}

		if testIsActiv {
			endTesting()
		} else {
			startTestint()

			reqSetting := &core.ReqSendingSettings{
				Requests:            activRequsts,
				Count_Workers:       10,
				Delay:               time.Duration(delaySlider.Value) * time.Millisecond,
				Duration:            time.Duration(durationSlider.Value) * time.Second,
				RequestChanBufSize:  10,
				ResponseChanBufSize: 10,
			}

			outChan := make(chan *core.RequestInfo, 10)
			go core.StartSendingHttpRequests(outChan, reqSetting, testCtx)

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

func endTesting() {

}
