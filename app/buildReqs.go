package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type RequestRow struct {
	method    *widget.Select
	url       *widget.Entry
	body      *widget.Entry
	container *fyne.Container
}

func showConfReqWindow(reqs *[]RequestRow, winOpen *bool) {
	confWindow := fyne.CurrentApp().NewWindow("Configure Requests")

	requestsContainer := container.NewVBox()

	createRequestRow := func() *fyne.Container {
		methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, nil)
		methodSelect.SetSelected("GET")
		methodSelect.Resize(fyne.NewSize(methodSelect.MinSize().Width, 30))

		urlEntry := widget.NewEntry()
		urlEntry.SetPlaceHolder("Enter URL (e.g. http://example.com)")
		urlEntry.Resize(fyne.NewSize(urlEntry.MinSize().Width, 30))

		bodyEntry := widget.NewMultiLineEntry()
		bodyEntry.SetPlaceHolder("Request body (optional)")
		bodyEntry.SetMinRowsVisible(1)
		bodyEntry.Resize(fyne.NewSize(bodyEntry.MinSize().Width, 30))

		row := container.NewGridWithColumns(3,
			methodSelect,
			urlEntry,
			bodyEntry,
		)
		return row
	}

	for _, req := range *reqs {
		methodSelect := req.method
		methodSelect.Resize(fyne.NewSize(methodSelect.MinSize().Width, 30))
		urlEntry := req.url
		urlEntry.Resize(fyne.NewSize(urlEntry.MinSize().Width, 30))
		bodyEntry := req.body
		bodyEntry.Resize(fyne.NewSize(bodyEntry.MinSize().Width, 30))

		row := container.NewGridWithColumns(3,
			methodSelect,
			urlEntry,
			bodyEntry,
		)
		requestsContainer.Add(row)
	}

	requestsContainer.Add(createRequestRow())

	addButton := widget.NewButton("Add Request", func() {
		if len(*reqs)+len(requestsContainer.Objects) >= 10 {
			dialog.ShowInformation("Error", "You can add a maximum of 10 requests", confWindow)
			return
		}
		requestsContainer.Add(createRequestRow())
	})
	applyButton := widget.NewButton("Apply", func() {
		*reqs = nil

		for _, obj := range requestsContainer.Objects {
			if row, ok := obj.(*fyne.Container); ok {
				methodSelect := row.Objects[0].(*widget.Select)
				urlEntry := row.Objects[1].(*widget.Entry)
				bodyEntry := row.Objects[2].(*widget.Entry)

				if methodSelect.Selected != "" && urlEntry.Text != "" {
					*reqs = append(*reqs, RequestRow{
						method: methodSelect,
						url:    urlEntry,
						body:   bodyEntry,
					})
				}
			}
		}
		*winOpen = false
		confWindow.Close()
	})

	confWindow.SetCloseIntercept(func() {
		*winOpen = false
		confWindow.Close()
	})

	content := container.NewBorder(addButton, applyButton, nil, nil, requestsContainer)

	confWindow.SetContent(content)
	confWindow.Resize(fyne.NewSize(600, 400))

	confWindow.SetOnClosed(func() {

	})

	confWindow.Show()
}
