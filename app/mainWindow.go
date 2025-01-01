package app

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
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
	REQ_DELAY_STEP             = 1.0
	TEST_DURATION_STEP         = 0.5
	UPDATE_SLIDERS_ENTRY_DELAY = 10 * time.Millisecond
)

var (
	// Main window
	window fyne.Window

	// Test button
	testButton  *widget.Button
	testIsActiv bool

	// Information about requests
	infoReqsGrid *widget.TextGrid

	// Stats label
	StatsLabel *widget.Label

	// Configurate requests
	configRequestsButton *widget.Button
	confWindowOpen       bool
	activRequstsRows     []*RequestRow
	activRequsts         []*http.Request

	// Protocol selection
	protocolButton   *widget.Button
	protocolSelect   *widget.Select
	selectedProtocol string

	// Report button
	reportButton     *widget.Button
	reportWindowOpen bool

	// Report info
	currentReports  []*core.RequestReport
	countReqs       atomic.Int64
	countFailedReqs atomic.Int64

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
	testCtx          context.Context
	testCancel       context.CancelFunc
	displayCtx       context.Context
	displayCtxCancel context.CancelFunc
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

	workersSlider = widget.NewSlider(1, float64(core.MAX_COUNT_WORKERS))
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

	// Stats label
	StatsLabel = widget.NewLabel("Time left: 00:00\nTime elapsed: 00:00\nRequests sent: 0\nRequests failed: 0")

	// OnChanged for sliders
	var timer *time.Timer

	updateUI := func(val string, updateFunc func(string)) {
		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(UPDATE_SLIDERS_ENTRY_DELAY, func() {
			updateFunc(val)
		})
	}

	delaySlider.OnChanged = func(f float64) {
		delayValStr = fmt.Sprintf("%v ms", f)
		updateUI(delayValStr, delayEntry.SetText)
	}
	durationSlider.OnChanged = func(f float64) {
		durationValStr = fmt.Sprintf("%v min", f)
		updateUI(durationValStr, durationEntry.SetText)
	}
	workersSlider.OnChanged = func(f float64) {
		workersValStr := fmt.Sprintf("%d", int(f))
		updateUI(workersValStr, workersEntry.SetText)
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
			if val > float64(core.MAX_COUNT_WORKERS) {
				val = float64(core.MAX_COUNT_WORKERS)
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
	displayCtx, displayCtxCancel = context.Background(), func() {}

	testButton = widget.NewButton("Start testing", testButtonFunc)

	reportButton = widget.NewButton("Show report", showReport)

	protocolButton = widget.NewButton("Change protocol", showProtocolWindow)
	selectedProtocol = core.DEFAULT_PROTO

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
		widget.NewLabel("Count of clients"),
		container.NewGridWrap(fyne.NewSize(300, 40), workersSlider),
		container.NewGridWrap(fyne.NewSize(70, 40), workersEntry),
	)

	// Wrap output in scroll container
	scrollOutput := container.NewScroll(infoReqsGrid)

	// Create left panel with fixed width
	leftPanel := container.NewVBox(
		widget.NewCard("Stats", "", container.NewVBox(
			StatsLabel,
		)),
		widget.NewCard("Testing options", "", container.NewVBox(
			showRequest,
			showBody,
			showHeaders,
			showTime,
		)),
		widget.NewCard("Settings", "", container.NewVBox(
			delayContainer,
			durationContainer,
			workersContainer,
			container.NewHBox(
				layout.NewSpacer(),
				protocolButton,
				layout.NewSpacer(),
				configRequestsButton,
				layout.NewSpacer(),
			),
		)),
	)

	// Bottom panel with fixed height
	bottomPanel := container.NewHBox(
		layout.NewSpacer(),
		reportButton,
		layout.NewSpacer(),
		testButton,
		layout.NewSpacer(),
	)

	mainContainer := container.NewBorder(
		nil,
		bottomPanel,
		leftPanel,
		nil,
		scrollOutput,
	)

	window.SetContent(mainContainer)
	window.Resize(fyne.NewSize(1200, 600))
	window.ShowAndRun()
}
