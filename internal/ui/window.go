package ui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var tabCounter int = 0

// NewMainWindow создает главное окно приложения
func NewMainWindow(fyneApp fyne.App) fyne.Window {
	window := fyneApp.NewWindow("Prospect - Система вкладок")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	// Создаем кастомный виджет табов с кнопками закрытия
	browserTabs := NewBrowserTabs()
	browserTabs.SetAddButtonCallback(func() {
		CreateTabWithClose(browserTabs)
	})

	// Создаем первую вкладку по умолчанию
	CreateTabWithClose(browserTabs)

	window.SetContent(browserTabs)
	return window
}

// AddTab добавляет новую вкладку в AppTabs
func AddTab(tabs *container.AppTabs, name string, content fyne.CanvasObject) {
	tabCounter++
	if name == "" {
		name = fmt.Sprintf("Вкладка #%d", tabCounter)
	}

	newTab := container.NewTabItem(name, content)
	tabs.Append(newTab)
	tabs.Select(newTab)
	fmt.Fprintf(os.Stdout, "[INFO] Добавлена вкладка: %s\n", name)
}

// CreateTab создает новую вкладку по умолчанию
func CreateTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Вкладка #%d", tabCounter)

	textArea := widget.NewMultiLineEntry()
	textArea.SetText("Новая вкладка\n\nВы можете редактировать этот текст.")
	textArea.Wrapping = fyne.TextWrapWord

	clearBtn := widget.NewButton("Очистить", func() {
		textArea.SetText("")
		fmt.Println("[INFO] Текст очищен")
	})

	resetBtn := widget.NewButton("Сбросить", func() {
		textArea.SetText("Новая вкладка\n\nВы можете редактировать этот текст.")
		fmt.Println("[INFO] Текст сброшен")
	})

	topContainer := container.NewVBox(
		widget.NewLabel("Редактор текста:"),
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
	// Не увеличиваем счетчик здесь, он увеличится в AddTab
	tabName := ""

	textArea := widget.NewMultiLineEntry()
	textArea.SetText("Новая вкладка\n\nВы можете редактировать этот текст.")
	textArea.Wrapping = fyne.TextWrapWord

	clearBtn := widget.NewButton("Очистить", func() {
		textArea.SetText("")
		fmt.Println("[INFO] Текст очищен")
	})

	resetBtn := widget.NewButton("Сбросить", func() {
		textArea.SetText("Новая вкладка\n\nВы можете редактировать этот текст.")
		fmt.Println("[INFO] Текст сброшен")
	})

	topContainer := container.NewVBox(
		widget.NewLabel("Редактор текста:"),
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
	browserTabs.AddTab(tabName, content)
}
