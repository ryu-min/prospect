package app

import (
	"fmt"
	"log"

	"prospect/internal/protobuf"
	"prospect/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App представляет основное приложение
type App struct {
	fyneApp fyne.App
	window  fyne.Window
}

// New создает новое приложение
func New() *App {
	fyneApp := app.New()
	log.Printf("Fyne application created successfully")

	return &App{
		fyneApp: fyneApp,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	// Проверяем наличие protoc
	log.Printf("Checking for protoc...")
	if err := protobuf.CheckProtoc(); err != nil {
		log.Printf("Error: %v", err)
		log.Printf("Please install protoc to work with protobuf files")
		log.Printf("You can install via: scoop install protobuf")
		return fmt.Errorf("protoc not found: %w", err)
	}

	// Создаем главное окно
	a.window = ui.NewMainWindow(a.fyneApp)
	log.Printf("Main window created")

	// Показываем окно и запускаем главный цикл
	log.Printf("Starting application main loop...")
	a.window.ShowAndRun()

	return nil
}

// GetApp возвращает экземпляр Fyne приложения
func (a *App) GetApp() fyne.App {
	return a.fyneApp
}
