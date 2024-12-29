package app

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showReport() {
	if len(currentReports) == 0 {
		dialog.ShowInformation("Error", "Before viewing the report, you must conduct a test", window)
		return
	}

	reportWindow := fyne.CurrentApp().NewWindow("Report")

	var reportText strings.Builder

	for _, reqsRep := range currentReports {
		reportText.WriteString(fmt.Sprintf("URL: %s\nAverage response time: %d ms\nMaximum response time: %d ms\nMinimal response time: %d ms\nNumber of requests: %d\n", reqsRep.Url, reqsRep.AvgTime.Milliseconds(), reqsRep.MaxTime.Milliseconds(), reqsRep.MinTime.Milliseconds(), reqsRep.Count))
		for code, count := range reqsRep.ReqCods {
			reportText.WriteString(fmt.Sprintf("Request code: %d, frequency: %d\n", code, count))
		}
		reportText.WriteString("Errors during request:\n")
		for err, count := range reqsRep.Errors {
			reportText.WriteString(fmt.Sprintf("%s : %d", err, count))
		}
		reportText.WriteRune('\n')
	}

	reportGrid := widget.NewTextGrid()
	reportGrid.SetText(reportText.String())

	content := container.NewVScroll(reportGrid)

	reportWindow.SetContent(content)
	reportWindow.Resize(fyne.NewSize(800, 600))

	reportWindow.Show()
}
