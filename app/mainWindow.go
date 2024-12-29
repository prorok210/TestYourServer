package app

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

var (
	// Main window
	window fyne.Window

	// Test button
	testButton  *widget.Button
	testIsActiv bool

	// Information about requests
	infoReqsGrid *widget.TextGrid

	// Configurate requests
	configRequestsButton *widget.Button
	confWindowOpen       bool
	activRequstsRows     []*RequestRow
	activRequsts         []*http.Request

	// Protocol selection
	protocolButton *widget.Button = widget.NewButton("Select protocol", func() {})

	// Report button
	reportButton *widget.Button

	// Report info
	currentReports []*core.RequestReport

	// Sliders
	delaySlider    *widget.Slider
	durationSlider *widget.Slider
	workersSlider  *widget.Slider

	// Entry for delay, duration and count of workers
	delayEntry    *widget.Entry
	durationEntry *widget.Entry
	workersEntry  *widget.Entry

	// Options for showing request
	showRequest *widget.Check
	showTime    *widget.Check
	showBody    *widget.Check
	showHeaders *widget.Check

	// Context for testing
	testCtx    context.Context
	testCancel context.CancelFunc
)

func CreateAppWindow() {
	a := app.New()
	window = a.NewWindow("Test Your Server")

	infoReqsGrid = widget.NewTextGrid()
	infoReqsGrid.SetText("There will be information about the requests here...")

	delaySlider = widget.NewSlider(float64(core.MIN_REQ_DELAY.Milliseconds()), float64(core.MAX_REQ_DELAY.Milliseconds()))
	delaySlider.Step = REQ_DELAY_STEP
	delaySlider.SetValue(float64(core.DEFAULT_REQ_DELAY.Milliseconds()))
	delayValStr := fmt.Sprintf("%v ms", core.DEFAULT_REQ_DELAY.Milliseconds())

	durationSlider = widget.NewSlider(float64(core.MIN_DURATION.Minutes()), float64(core.MAX_DURATION.Minutes()))
	durationSlider.Step = TEST_DURATION_STEP
	durationSlider.SetValue(float64(core.DEFAULT_DURATION.Minutes()))
	durationValStr := fmt.Sprintf("%v min", core.DEFAULT_DURATION.Minutes())

	workersSlider = widget.NewSlider(1, float64(core.MAX_CCOUNT_WORKERS))
	workersSlider.Step = 1
	workersSlider.SetValue(core.DEFAULT_COUNT_WORKERS)

	delayEntry = widget.NewEntry()
	delayEntry.SetText(delayValStr)
	delayEntry.Resize(fyne.NewSize(100, 1000))

	durationEntry = widget.NewEntry()
	durationEntry.SetText(durationValStr)
	durationEntry.Resize(fyne.NewSize(100, 1000))

	workersEntry = widget.NewEntry()
	workersEntry.SetText(fmt.Sprintf("%v", core.DEFAULT_COUNT_WORKERS))
	workersEntry.Resize(fyne.NewSize(100, 1000))

	// OnChanged for sliders
	delaySlider.OnChanged = func(f float64) {
		delayValStr = fmt.Sprintf("%v ms", f)
		delayEntry.SetText(delayValStr)
	}
	durationSlider.OnChanged = func(f float64) {
		durationValStr = fmt.Sprintf("%v min", f)
		durationEntry.SetText(durationValStr)
	}
	workersSlider.OnChanged = func(f float64) {
		workersEntry.SetText(fmt.Sprintf("%v", f))
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

	workersEntry.OnChanged = func(s string) {
		matched, _ := regexp.MatchString(`^[0-9]+$`, s)

		if matched {
			val, _ := strconv.ParseFloat(s, 64)
			if val > float64(core.MAX_CCOUNT_WORKERS) {
				val = float64(core.MAX_CCOUNT_WORKERS)
			}
			if val < 1 {
				val = 1
			}
			workersSlider.SetValue(val)
			workersEntry.SetText(fmt.Sprintf("%v", val))
		} else {
			workersEntry.SetText(fmt.Sprintf("%v", 1))
		}
	}

	showRequest = widget.NewCheck("Show request", nil)
	showTime = widget.NewCheck("Show response Time", nil)
	showBody = widget.NewCheck("Show response Body (only first 1000 bytes)", nil)
	showHeaders = widget.NewCheck("Show response Headers (only first 10 headers)", nil)

	testCtx, testCancel = context.Background(), func() {}

	testButton = widget.NewButton("Start testing", testButtonFunc)

	reportButton = widget.NewButton("Show report", showReport)

	configRequestsButton = widget.NewButton("Configurate requests", func() {
		if testIsActiv {
			dialog.ShowInformation("Error", "You can't configure requests while testing is in progress", window)
			return
		}
		if !confWindowOpen {
			showConfReqWindow(&activRequstsRows, &activRequsts, &confWindowOpen)
			confWindowOpen = true
		}
	})

	window.SetCloseIntercept(func() {
		if confWindowOpen {
			dialog.ShowInformation("Info", "Please close the settings window before exiting.", window)
		} else {
			window.Close()
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
	workersContainer := container.NewHBox(
		widget.NewLabel("Count of workers"),
		container.NewGridWrap(fyne.NewSize(300, 40), workersSlider),
		container.NewGridWrap(fyne.NewSize(70, 40), workersEntry),
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
			workersContainer,
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

	window.SetContent(content)

	window.Resize(fyne.NewSize(1200, 600))

	window.ShowAndRun()
}
