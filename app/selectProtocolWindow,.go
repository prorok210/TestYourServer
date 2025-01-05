package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/prorok210/TestYourServer/core"
)

var (
	protocolWindowOpen bool
	protocolSelect     *widget.Select
	selectedProtocol   core.Protocol
	secureCheck        *widget.Check
	disableCheckTls    bool
)

func showProtocolWindow() {
	if protocolWindowOpen {
		return
	}
	protocolWindowOpen = true

	protocolWindow := fyne.CurrentApp().NewWindow("Select protocol")

	protocolOptions := []string{"HTTP", "WS"}

	protocolSelect = widget.NewSelect(protocolOptions, func(s string) {
		if s != selectedProtocol.String() {
			activRequsts = []core.Request{}
			activRequstsRows = []*RequestRow{}
			requestsContainer.Objects = []fyne.CanvasObject{}
			requestsContainer.Add(createRequestRow())
		}
		switch s {
		case "HTTP":
			selectedProtocol = core.HTTP
		case "WS":
			selectedProtocol = core.WS
		}
	})

	protocolSelect.SetSelected(selectedProtocol.String())

	secureCheck = widget.NewCheck("Disable TLS checking", func(b bool) {
		disableCheckTls = b
	})
	secureCheck.SetChecked(disableCheckTls)

	protocolWindow.SetContent(
		container.NewVBox(
			widget.NewLabel("Select protocol"),
			protocolSelect,
			secureCheck,
			widget.NewButton("OK", func() {
				switch protocolSelect.Selected {
				case "HTTP":
					selectedProtocol = core.HTTP
				case "WS":
					selectedProtocol = core.WS
				}
				protocolWindow.Close()
				protocolWindowOpen = false
			}),
		),
	)

	protocolWindow.SetOnClosed(func() {
		protocolWindowOpen = false
	})

	protocolWindow.Show()
}
