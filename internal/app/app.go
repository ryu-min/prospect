package app

import (
	"fmt"

	"prospect/internal/protobuf"
	"prospect/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window
}

func New() *App {
	fyneApp := app.New()
	return &App{
		fyneApp: fyneApp,
	}
}

func (a *App) Run() error {
	if err := protobuf.CheckProtoc(); err != nil {
		return fmt.Errorf("protoc not found: %w", err)
	}
	a.window = ui.NewMainWindow(a.fyneApp)
	a.window.ShowAndRun()
	return nil
}

func (a *App) GetApp() fyne.App {
	return a.fyneApp
}
