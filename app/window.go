package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/prorok210/TestYourServer/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateAppWindow() {
	a := app.New()
	w := a.NewWindow("Test Your Server")

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("There will be information about the request here...")

	scrollContainer := container.NewScroll(entry)

	ctx, cancel := context.WithCancel(context.Background())

	var isTesting bool
	var testButton *widget.Button

	showTime := widget.NewCheck("Show Time", nil)
	showBody := widget.NewCheck("Show Body (only first 1000 bytes)", nil)
	showHeaders := widget.NewCheck("Show Headers", nil)

	testButton = widget.NewButton("Start testing", func() {
		if isTesting {
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			testButton.SetText("Start testing")
			isTesting = false
		} else {
			outChan := make(chan *core.RequestInfo, 10)
			go core.StartSendingHttpRequests(outChan, ctx)

			go func() {
				var lastRequests []string
				maxLines := 50
				for {
					select {
					case <-ctx.Done():
						return
					case resp := <-outChan:
						if resp != nil && resp.Response != nil {
							var responseText string
							if showTime.Checked {
								responseText += fmt.Sprintf("Time: %v\n", resp.Time)
							}
							responseText += fmt.Sprintf("Status: %s\n", resp.Response.Status)

							if showBody.Checked {
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
							}
							if showHeaders.Checked {
								responseText += "Headers:\n"
								for k, v := range resp.Response.Header {
									responseText += fmt.Sprintf("%s: %s\n", k, v)
								}
							}

							lastRequests = append(lastRequests, responseText)
							if len(lastRequests) > maxLines {
								lastRequests = lastRequests[1:]
							}

							entry.SetText(strings.Join(lastRequests, "\n"))
						} else {
							entry.SetText(entry.Text + "Error with get response\n")
						}
					}
				}
			}()
			testButton.SetText("Stop testing")
			isTesting = true
		}
	})

	optionsContainer := container.NewVBox(showTime, showBody, showHeaders)

	w.SetContent(container.NewBorder(container.NewVBox(testButton, optionsContainer), nil, nil, nil, scrollContainer))

	w.Resize(fyne.NewSize(1000, 1000))
	w.ShowAndRun()
}
