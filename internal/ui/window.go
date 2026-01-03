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

	// Создаем контейнер вкладок напрямую
	tabs := container.NewAppTabs()

	// Создаем панель управления как первую вкладку (без увеличения счетчика)
	controlTab := createControlTab(tabs)
	controlTabItem := container.NewTabItem("Панель управления", controlTab)
	tabs.Append(controlTabItem)

	// Создаем панель инструментов
	toolbar := createToolbar(tabs)

	// Используем Border layout: toolbar сверху, tabs в центре
	content := container.NewBorder(
		toolbar, // верх
		nil,     // низ
		nil,     // лево
		nil,     // право
		tabs,    // центр
	)

	window.SetContent(content)
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

	// Добавляем вкладку без увеличения счетчика, так как он уже увеличен
	newTab := container.NewTabItem(tabName, content)
	tabs.Append(newTab)
	tabs.Select(newTab)
	fmt.Fprintf(os.Stdout, "[INFO] Добавлена вкладка: %s\n", tabName)
}

// createControlTab создает вкладку с панелью управления
func createControlTab(tabs *container.AppTabs) fyne.CanvasObject {
	title := widget.NewLabel("Система управления вкладками")
	title.TextStyle = fyne.TextStyle{Bold: true}

	infoLabel := widget.NewLabel("Нажмите на кнопку, чтобы создать новую вкладку")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Кнопка для создания новой вкладки
	createBtn := widget.NewButton("Создать новую вкладку", func() {
		CreateTab(tabs)
	})
	createBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		infoLabel,
		widget.NewSeparator(),
		createBtn,
	)

	return container.NewScroll(content)
}

// createToolbar создает панель инструментов
func createToolbar(tabs *container.AppTabs) fyne.CanvasObject {
	addButton := widget.NewButton("+ Добавить вкладку", func() {
		CreateTab(tabs)
	})
	addButton.Importance = widget.HighImportance

	toolbar := container.NewHBox(
		addButton,
		widget.NewSeparator(),
		widget.NewLabel("Доступно из любого таба"),
	)

	return container.NewPadded(toolbar)
}
