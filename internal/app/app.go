package app

import (
	"fmt"
	"os"

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
	fmt.Fprintln(os.Stdout, "[INFO] Приложение Fyne создано успешно")

	return &App{
		fyneApp: fyneApp,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	// Проверяем наличие protoc
	fmt.Fprintln(os.Stdout, "[INFO] Проверка наличия protoc...")
	if err := protobuf.CheckProtoc(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		fmt.Fprintln(os.Stderr, "[ERROR] Установите protoc для работы с protobuf файлами")
		fmt.Fprintln(os.Stderr, "[ERROR] Можно установить через: scoop install protobuf")
		return fmt.Errorf("protoc не найден: %w", err)
	}

	// Создаем главное окно
	a.window = ui.NewMainWindow(a.fyneApp)
	fmt.Fprintln(os.Stdout, "[INFO] Главное окно создано")

	// Показываем окно и запускаем главный цикл
	fmt.Fprintln(os.Stdout, "[INFO] Запуск главного цикла приложения...")
	a.window.ShowAndRun()

	return nil
}

// GetApp возвращает экземпляр Fyne приложения
func (a *App) GetApp() fyne.App {
	return a.fyneApp
}
