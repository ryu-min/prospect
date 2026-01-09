package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var tabCounter int = 0
var undefinedTabCounter int = 0

func NewMainWindow(fyneApp fyne.App) fyne.Window {
	window := fyneApp.NewWindow("prospect")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	icon := createAppIcon()
	if icon != nil {
		window.SetIcon(icon)
	}

	browserTabs := newTabManager()
	browserTabs.SetAddButtonCallback(func() {
		createTabWithClose(browserTabs)
	})

	if err := loadTabState(browserTabs, fyneApp, window); err != nil {
		createProtoTab(browserTabs, fyneApp, window)
	} else if len(browserTabs.tabs) == 0 {
		createProtoTab(browserTabs, fyneApp, window)
	}

	window.SetOnClosed(func() {
		saveTabState(browserTabs)
	})

	toolbar := browserTabs.GetToolbarManager().GetToolbar()
	mainContent := container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		browserTabs,
	)

	window.SetContent(mainContent)
	return window
}

func createTabWithClose(browserTabs *tabManager) {
	var fyneApp fyne.App
	var parentWindow fyne.Window

	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		parentWindow = windows[0]
		fyneApp = fyne.CurrentApp()
	} else {
		fyneApp = fyne.CurrentApp()
		parentWindow = fyneApp.NewWindow("")
	}

	content := protoView(fyneApp, parentWindow, browserTabs)

	undefinedTabCounter++
	tabTitle := fmt.Sprintf("undefined_%d", undefinedTabCounter)
	browserTabs.AddTab(tabTitle, content)
}

func createProtoTab(browserTabs *tabManager, fyneApp fyne.App, parentWindow fyne.Window) {
	content := protoView(fyneApp, parentWindow, browserTabs)
	undefinedTabCounter++
	tabTitle := fmt.Sprintf("undefined_%d", undefinedTabCounter)
	browserTabs.AddTab(tabTitle, content)
}
