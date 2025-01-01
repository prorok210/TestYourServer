package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

const (
	MAX_URL_LEN = 100
	MAX_ROW_LEN = 100
)

func showReport() {
	if reportWindowOpen {
		return
	}

	reportWindow := fyne.CurrentApp().NewWindow("Report")

	reportWindowOpen = true

	reportWindow.SetOnClosed(func() {
		reportWindowOpen = false
		reportWindow.Close()
	})

	sections := []fyne.CanvasObject{}

	for _, reqsRep := range currentReports {
		urlLabel := widget.NewLabelWithStyle(
			core.TruncateString(fmt.Sprintf("URL: %s", reqsRep.Url), MAX_URL_LEN),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)

		info := widget.NewLabel(fmt.Sprintf(
			"Average response time: %d ms\nMaximum response time: %d ms\nMinimal response time: %d ms\nNumber of requests: %d",
			reqsRep.AvgTime.Milliseconds(),
			reqsRep.MaxTime.Milliseconds(),
			reqsRep.MinTime.Milliseconds(),
			reqsRep.Count,
		))

		reqCodes := widget.NewLabel("Request codes and frequencies:")
		reqCodeContent := ""
		for code, count := range reqsRep.ReqCods {
			reqCodeContent += fmt.Sprintf("  - Code: %d, Frequency: %d\n", code, count)
		}
		reqCodesContent := widget.NewLabel(reqCodeContent)

		errorsLabel := widget.NewLabelWithStyle("Errors during requests:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		errorsContent := ""
		if len(reqsRep.Errors) == 0 {
			errorsContent = "No errors.\n"
		} else {
			for err, count := range reqsRep.Errors {
				wrappedError := core.WrapText(fmt.Sprintf("  - %s: %d", err, count), MAX_ROW_LEN)
				fmt.Println("wrappedError", wrappedError)
				errorsContent += wrappedError + "\n"
			}
		}
		errorsContentLabel := widget.NewLabel(errorsContent)

		sections = append(sections, container.NewVBox(
			urlLabel,
			info,
			reqCodes,
			reqCodesContent,
			errorsLabel,
			errorsContentLabel,
			widget.NewSeparator(),
		))
	}

	reportContent := container.NewVScroll(container.NewVBox(sections...))

	if len(currentReports) == 0 {
		reportContent = container.NewVScroll(widget.NewLabel("No reports."))
	}

	reportWindow.SetContent(reportContent)
	reportWindow.Resize(fyne.NewSize(800, 600))
	reportWindow.Show()
}
