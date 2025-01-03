package app

import (
	"fmt"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

const (
	MAX_COUNT_REQS = 100
)

var (
	protocolButton    *widget.Button
	confWindowOpen    bool
	activRequstsRows  []*RequestRow
	activRequsts      []core.Request
	requestsContainer *fyne.Container
)

type RequestRow struct {
	method    *widget.Select
	url       *widget.Entry
	body      *widget.Entry
	delete    *widget.Button
	container *fyne.Container
}

func createRequestRow(deleteRow func(*fyne.Container)) *fyne.Container {
	var row *fyne.Container
	methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, nil)
	methodSelect.SetSelected("GET")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL (e.g. http://example.com)")

	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetPlaceHolder("Request body (optional)")
	bodyEntry.SetMinRowsVisible(1)
	bodyEntry.OnChanged = func(s string) {
		if s == "" {
			bodyEntry.SetPlaceHolder("Request body (optional)")
			bodyEntry.SetMinRowsVisible(1)
		} else {
			bodyEntry.SetPlaceHolder("")
			if len(strings.Split(s, "\n")) < 3 {
				bodyEntry.SetMinRowsVisible(3)
			} else {
				bodyEntry.SetMinRowsVisible(len(strings.Split(s, "\n")))
			}
		}
	}

	deleteButton := widget.NewButton("❌", func() {
		deleteRow(row)
	})

	split1 := container.NewHSplit(methodSelect, urlEntry)
	split1.Offset = 0.01
	split2 := container.NewHSplit(bodyEntry, deleteButton)
	split2.Offset = 0.99

	row = container.NewAdaptiveGrid(1,
		container.NewHSplit(
			split1,
			split2,
		),
	)

	return row
}

func deleteRow(row *fyne.Container) {
	for i, r := range activRequstsRows {
		if r.container == row {
			activRequstsRows = append((activRequstsRows)[:i], (activRequstsRows)[i+1:]...)
			break
		}
	}
	requestsContainer.Remove(row)
}

func showConfReqWindow() {
	confWindow := fyne.CurrentApp().NewWindow("Configure Requests")

	requestsContainer = container.NewVBox()

	if len(activRequstsRows) == 0 {
		requestsContainer.Add(createRequestRow(deleteRow))
	}

	for _, req := range activRequstsRows {
		methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, nil)
		methodSelect.SetSelected(req.method.Selected)

		urlEntry := widget.NewEntry()
		urlEntry.SetText(strings.TrimSpace(req.url.Text))

		bodyEntry := widget.NewMultiLineEntry()
		bodyEntry.SetText(req.body.Text)
		if req.body.Text == "" {
			bodyEntry.SetPlaceHolder("Request body (optional)")
			bodyEntry.SetMinRowsVisible(1)
		} else {
			bodyEntry.SetPlaceHolder("")
			if len(strings.Split(req.body.Text, "\n")) < 3 {
				bodyEntry.SetMinRowsVisible(3)
			} else {
				bodyEntry.SetMinRowsVisible(len(strings.Split(req.body.Text, "\n")))
			}
		}

		bodyEntry.OnChanged = func(s string) {
			if s == "" {
				bodyEntry.SetPlaceHolder("Request body (optional)")
				bodyEntry.SetMinRowsVisible(1)
			} else {
				bodyEntry.SetPlaceHolder("")
				if len(strings.Split(s, "\n")) < 3 {
					bodyEntry.SetMinRowsVisible(3)
				} else {
					bodyEntry.SetMinRowsVisible(len(strings.Split(s, "\n")))
				}
			}
		}

		var row *fyne.Container

		deleteButton := widget.NewButton("❌", func() {
			deleteRow(row)
		})

		split1 := container.NewHSplit(methodSelect, urlEntry)
		split1.Offset = 0.01
		split2 := container.NewHSplit(bodyEntry, deleteButton)
		split2.Offset = 0.99

		row = container.NewAdaptiveGrid(1,
			container.NewHSplit(
				split1,
				split2,
			),
		)

		req.method = methodSelect
		req.url = urlEntry
		req.body = bodyEntry
		req.delete = deleteButton
		req.container = row

		requestsContainer.Add(row)
	}

	addButton := widget.NewButton("Add Request", func() {
		if len(requestsContainer.Objects) >= MAX_COUNT_REQS {
			dialog.ShowInformation("Error", fmt.Sprintf("You can add a maximum of %d requests", MAX_COUNT_REQS), confWindow)
			return
		}
		requestsContainer.Add(createRequestRow(deleteRow))
	})

	clearButton := widget.NewButton("Clear", func() {
		requestsContainer.Objects = nil
		requestsContainer.Add(createRequestRow(deleteRow))
	})

	applyButton := widget.NewButton("Ok", func() {
		if protocolWindowOpen {
			dialog.ShowInformation("Info", "Please close the settings window before exiting.", confWindow)
			return
		}

		var err error

		defer func() {
			confWindowOpen = false
			if err == nil {
				confWindow.Close()
			}
		}()

		activRequstsRows = nil
		activRequsts = nil

		for _, obj := range requestsContainer.Objects {
			if row, ok := obj.(*fyne.Container); ok {
				hSplit, ok := row.Objects[0].(*container.Split)
				if !ok {
					continue
				}
				split1, ok := hSplit.Leading.(*container.Split)
				if !ok {
					continue
				}

				split2, ok := hSplit.Trailing.(*container.Split)
				if !ok {
					continue
				}

				methodSelect, ok := split1.Leading.(*widget.Select)
				if !ok {
					continue
				}

				urlEntry, ok := split1.Trailing.(*widget.Entry)
				if !ok {
					continue
				}

				bodyEntry, ok := split2.Leading.(*widget.Entry)
				if !ok {
					continue
				}

				deleteButton, ok := split2.Trailing.(*widget.Button)
				if !ok {
					continue
				}

				if methodSelect.Selected != "" && urlEntry.Text != "" {
					err = core.ValidateURL(urlEntry.Text, selectedProtocol)
					if err != nil {
						dialog.ShowInformation("Error", err.Error(), confWindow)
						return
					}

					activRequstsRows = append(activRequstsRows, &RequestRow{
						method:    methodSelect,
						url:       urlEntry,
						body:      bodyEntry,
						delete:    deleteButton,
						container: row,
					})

					var newReq core.Request

					switch selectedProtocol {
					case core.HTTP:
						req, err := http.NewRequest(methodSelect.Selected, urlEntry.Text, strings.NewReader(bodyEntry.Text))
						if err != nil {
							dialog.ShowInformation("Error", "Invalid request", confWindow)
							return
						}
						newReq = &core.HTTPRequest{
							Request:    req,
							CachedBody: []byte(bodyEntry.Text),
						}
					case core.WS:
						newReq = &core.WSRequest{
							URI:     urlEntry.Text,
							Payload: []byte(bodyEntry.Text),
						}
					default:
						dialog.ShowInformation("Error", "Invalid protocol", confWindow)
						return
					}

					activRequsts = append(activRequsts, newReq)
				}
			}
		}
	})

	confWindow.SetCloseIntercept(func() {
		if protocolWindowOpen {
			dialog.ShowInformation("Info", "Please close the settings window before exiting.", confWindow)
			return
		}
		confWindowOpen = false
		confWindow.Close()
	})

	content := container.NewBorder(
		nil,
		container.NewVBox(container.NewAdaptiveGrid(2, clearButton, addButton), protocolButton, applyButton),
		nil,
		nil,
		container.NewVScroll(requestsContainer),
	)

	confWindow.SetContent(content)
	confWindow.Resize(fyne.NewSize(800, 600))

	confWindow.Show()
}
