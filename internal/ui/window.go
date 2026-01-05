package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var tabCounter int = 0
var undefinedTabCounter int = 0

// NewMainWindow создает главное окно приложения
func NewMainWindow(fyneApp fyne.App) fyne.Window {
	window := fyneApp.NewWindow("prospect")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	// Устанавливаем иконку приложения
	icon := createAppIcon()
	if icon != nil {
		window.SetIcon(icon)
	}

	// Создаем кастомный виджет табов с кнопками закрытия
	browserTabs := NewBrowserTabs()
	browserTabs.SetAddButtonCallback(func() {
		CreateTabWithClose(browserTabs)
	})

	// Создаем первую вкладку по умолчанию - Protobuf Viewer
	// Передаем окно для диалогов
	CreateProtobufTab(browserTabs, fyneApp, window)

	window.SetContent(browserTabs)
	return window
}

// AddTab добавляет новую вкладку в AppTabs
func AddTab(tabs *container.AppTabs, name string, content fyne.CanvasObject) {
	tabCounter++
	if name == "" {
		name = "Tab"
	}

	newTab := container.NewTabItem(name, content)
	tabs.Append(newTab)
	tabs.Select(newTab)
	log.Printf("Tab added: %s", name)
}

// CreateTab создает новую вкладку по умолчанию
func CreateTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := "Tab"

	textArea := widget.NewMultiLineEntry()
	textArea.SetText("Новая вкладка\n\nВы можете редактировать этот текст.")
	textArea.Wrapping = fyne.TextWrapWord

	clearBtn := widget.NewButton("Clear", func() {
		textArea.SetText("")
	})

	resetBtn := widget.NewButton("Reset", func() {
		textArea.SetText("New tab\n\nYou can edit this text.")
		log.Printf("Text reset")
	})

	topContainer := container.NewVBox(
		widget.NewLabel("Text editor:"),
	)

	bottomContainer := container.NewHBox(
		clearBtn,
		resetBtn,
	)

	content := container.NewBorder(
		topContainer,
		bottomContainer,
		nil,
		nil,
		container.NewScroll(textArea),
	)

	// Добавляем вкладку
	AddTab(tabs, tabName, content)
}

// CreateTabWithClose создает новую вкладку с кнопкой закрытия через BrowserTabs
func CreateTabWithClose(browserTabs *BrowserTabs) {
	// Получаем главное окно из приложения
	// Нужно передать fyneApp и window, но пока используем глобальный доступ
	// Для этого нужно изменить сигнатуру или использовать другой подход
	// Пока создаем protobuf view без окна (будет использоваться для диалогов)

	// Получаем приложение из первого окна
	var fyneApp fyne.App
	var parentWindow fyne.Window

	// Пытаемся получить окно из драйвера
	// Это не идеально, но работает
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		parentWindow = windows[0]
		fyneApp = fyne.CurrentApp()
	} else {
		// Если окна нет, создаем новое (не должно произойти)
		fyneApp = fyne.CurrentApp()
		parentWindow = fyneApp.NewWindow("")
	}

	// Создаем protobuf viewer
	content := ProtobufView(fyneApp, parentWindow, browserTabs)

	// Добавляем вкладку с undefined_{number}
	undefinedTabCounter++
	tabTitle := fmt.Sprintf("undefined_%d", undefinedTabCounter)
	browserTabs.AddTab(tabTitle, content)
}

// CreateProtobufTab создает вкладку для просмотра protobuf файлов
func CreateProtobufTab(browserTabs *BrowserTabs, fyneApp fyne.App, parentWindow fyne.Window) {
	content := ProtobufView(fyneApp, parentWindow, browserTabs)
	// Используем undefined_{number} если файл не открыт
	undefinedTabCounter++
	tabTitle := fmt.Sprintf("undefined_%d", undefinedTabCounter)
	browserTabs.AddTab(tabTitle, content)
}
