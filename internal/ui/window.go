package ui

import (
	"prospect/internal/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewMainWindow создает главное окно приложения
func NewMainWindow(fyneApp fyne.App) fyne.Window {
	window := fyneApp.NewWindow("Prospect - Система вкладок")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	// Создаем контейнер вкладок напрямую
	tabs := container.NewAppTabs()

	// Создаем панель управления как первую вкладку
	controlTab := createControlTab(tabs)
	widgets.AddTab(tabs, "Панель управления", controlTab)

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

// createControlTab создает вкладку с панелью управления
func createControlTab(tabs *container.AppTabs) fyne.CanvasObject {
	title := widget.NewLabel("Система управления вкладками")
	title.TextStyle = fyne.TextStyle{Bold: true}

	infoLabel := widget.NewLabel("Нажмите на кнопку, чтобы создать новую вкладку с виджетами")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Список доступных типов вкладок
	tabTypes := []struct {
		name string
		fn   func(*container.AppTabs)
	}{
		{"Текст", widgets.CreateTextTab},
		{"Форма", widgets.CreateFormTab},
		{"Список", widgets.CreateListTab},
		{"Ввод", widgets.CreateInputTab},
		{"Прогресс", widgets.CreateProgressTab},
		{"Кастомная", widgets.CreateCustomTab},
	}

	// Создаем кнопки для каждого типа
	buttonsContainer := container.NewVBox()
	for _, tabType := range tabTypes {
		tabType := tabType // capture для замыкания
		btn := widget.NewButton("Создать вкладку: "+tabType.name, func() {
			tabType.fn(tabs)
		})
		buttonsContainer.Add(btn)
	}

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		infoLabel,
		widget.NewSeparator(),
		buttonsContainer,
	)

	return container.NewScroll(content)
}

// createToolbar создает панель инструментов
func createToolbar(tabs *container.AppTabs) fyne.CanvasObject {
	// Список доступных типов вкладок
	tabTypes := []string{
		"Текст",
		"Форма",
		"Список",
		"Ввод",
		"Прогресс",
		"Кастомная",
	}

	// Маппинг типов на функции создания
	typeMap := map[string]func(*container.AppTabs){
		"Текст":     widgets.CreateTextTab,
		"Форма":     widgets.CreateFormTab,
		"Список":    widgets.CreateListTab,
		"Ввод":      widgets.CreateInputTab,
		"Прогресс":  widgets.CreateProgressTab,
		"Кастомная": widgets.CreateCustomTab,
	}

	tabTypeSelect := widget.NewSelect(tabTypes, func(selected string) {
		// Выбор типа не создает вкладку автоматически
	})

	addButton := widget.NewButton("+ Добавить вкладку", func() {
		selected := tabTypeSelect.Selected
		if selected == "" {
			// Если ничего не выбрано, создаем кастомную вкладку по умолчанию
			widgets.CreateCustomTab(tabs)
			return
		}

		// Создаем вкладку выбранного типа
		if createFn, exists := typeMap[selected]; exists {
			createFn(tabs)
		}
	})
	addButton.Importance = widget.HighImportance

	toolbar := container.NewHBox(
		widget.NewLabel("Тип:"),
		tabTypeSelect,
		addButton,
		widget.NewSeparator(),
		widget.NewLabel("Доступно из любого таба"),
	)

	return container.NewPadded(toolbar)
}
