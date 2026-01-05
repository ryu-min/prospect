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
	log.Printf("Приложение Fyne создано успешно")

	return &App{
		fyneApp: fyneApp,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	// Проверяем наличие protoc
	log.Printf("Проверка наличия protoc...")
	if err := protobuf.CheckProtoc(); err != nil {
		log.Printf("Ошибка: %v", err)
		log.Printf("Установите protoc для работы с protobuf файлами")
		log.Printf("Можно установить через: scoop install protobuf")
		return fmt.Errorf("protoc не найден: %w", err)
	}

	// Создаем главное окно
	a.window = ui.NewMainWindow(a.fyneApp)
	log.Printf("Главное окно создано")

	// Показываем окно и запускаем главный цикл
	log.Printf("Запуск главного цикла приложения...")
	a.window.ShowAndRun()

	return nil
}

// GetApp возвращает экземпляр Fyne приложения
func (a *App) GetApp() fyne.App {
	return a.fyneApp
}
