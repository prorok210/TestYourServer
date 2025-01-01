package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func showProtocolWindow() {
	protocolWindow := fyne.CurrentApp().NewWindow("Select protocol")

	protocols := []string{"HTTPS", "HTTP", "WebSocket"}

	protocolSelect = widget.NewSelect(protocols, func(s string) {
		selectedProtocol = s
	})

	if selectedProtocol == "" {
		selectedProtocol = "HTTPS"
	}

	protocolSelect.SetSelected(selectedProtocol)

	protocolWindow.SetContent(container.NewVBox(
		widget.NewLabel("Select protocol"),
		protocolSelect,
		widget.NewButton("OK", func() {
			selectedProtocol = protocolSelect.Selected
			protocolWindow.Close()
		}),
	))
	protocolWindow.Show()
}
