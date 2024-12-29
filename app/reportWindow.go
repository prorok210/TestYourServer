package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

const (
	MAX_URL_LEN = 100
)

func showReport() {
	if len(currentReports) == 0 {
		dialog.ShowInformation("Error", "Before viewing the report, you must conduct a test", window)
		return
	}

	if testIsActiv {
		dialog.ShowInformation("Error", "You can't open a report during a test", window)
		return
	}

	reportWindow := fyne.CurrentApp().NewWindow("Report")

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
				errorsContent += fmt.Sprintf("  - %s: %d\n", err, count)
			}
		}
		errorsContentLabel := widget.NewRichTextFromMarkdown(errorsContent)

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

	reportWindow.SetContent(reportContent)
	reportWindow.Resize(fyne.NewSize(800, 600))
	reportWindow.Show()
}
