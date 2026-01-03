package widgets

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var tabCounter int

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

// CreateTextTab создает вкладку с текстовыми виджетами
func CreateTextTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Текст #%d", tabCounter)

	label1 := widget.NewLabel("Это обычный текст")
	label2 := widget.NewLabel("Это многострочный текст.\nВторая строка.\nТретья строка.")
	label2.Wrapping = fyne.TextWrapWord

	richText := widget.NewRichTextFromMarkdown("# Заголовок\n\nЭто **жирный** текст и это *курсив*.\n\n- Пункт 1\n- Пункт 2\n- Пункт 3")

	content := container.NewVBox(
		label1,
		widget.NewSeparator(),
		label2,
		widget.NewSeparator(),
		richText,
	)

	AddTab(tabs, tabName, container.NewScroll(content))
}

// CreateFormTab создает вкладку с формой
func CreateFormTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Форма #%d", tabCounter)

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Введите имя")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Введите email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Введите пароль")

	checkBox := widget.NewCheck("Согласен с условиями", func(checked bool) {
		fmt.Printf("[INFO] Чекбокс: %v\n", checked)
	})

	radioGroup := widget.NewRadioGroup([]string{"Вариант 1", "Вариант 2", "Вариант 3"}, func(selected string) {
		fmt.Printf("[INFO] Выбран: %s\n", selected)
	})

	statusLabel := widget.NewLabel("")
	statusLabel.Importance = widget.HighImportance

	submitBtn := widget.NewButton("Отправить", func() {
		fmt.Printf("[INFO] Форма отправлена: имя=%s, email=%s\n", nameEntry.Text, emailEntry.Text)
		statusLabel.SetText("✓ Форма успешно отправлена!")
	})

	form := container.NewVBox(
		widget.NewLabel("Имя:"),
		nameEntry,
		widget.NewLabel("Email:"),
		emailEntry,
		widget.NewLabel("Пароль:"),
		passwordEntry,
		checkBox,
		radioGroup,
		statusLabel,
		submitBtn,
	)

	AddTab(tabs, tabName, container.NewScroll(form))
}

// CreateListTab создает вкладку со списком
func CreateListTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Список #%d", tabCounter)

	listData := []string{
		"Элемент 1", "Элемент 2", "Элемент 3", "Элемент 4", "Элемент 5",
		"Элемент 6", "Элемент 7", "Элемент 8", "Элемент 9", "Элемент 10",
	}

	list := widget.NewList(
		func() int {
			return len(listData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(listData[id])
		},
	)

	selectedLabel := widget.NewLabel("Выберите элемент из списка")

	list.OnSelected = func(id widget.ListItemID) {
		selectedLabel.SetText(fmt.Sprintf("Выбран: %s", listData[id]))
		fmt.Printf("[INFO] Выбран элемент: %s\n", listData[id])
	}

	content := container.NewBorder(
		widget.NewLabel("Список элементов:"),
		selectedLabel,
		nil,
		nil,
		list,
	)

	AddTab(tabs, tabName, content)
}

// CreateInputTab создает вкладку с элементами ввода
func CreateInputTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Ввод #%d", tabCounter)

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("Текстовое поле")

	multilineEntry := widget.NewMultiLineEntry()
	multilineEntry.SetPlaceHolder("Многострочное поле")
	multilineEntry.Wrapping = fyne.TextWrapWord

	slider := widget.NewSlider(0, 100)
	slider.SetValue(50)
	sliderValue := widget.NewLabel(fmt.Sprintf("Значение: %.0f", slider.Value))

	slider.OnChanged = func(value float64) {
		sliderValue.SetText(fmt.Sprintf("Значение: %.0f", value))
		fmt.Printf("[INFO] Слайдер изменен: %.0f\n", value)
	}

	content := container.NewVBox(
		widget.NewLabel("Текстовое поле:"),
		textEntry,
		widget.NewLabel("Многострочное поле:"),
		multilineEntry,
		widget.NewLabel("Слайдер:"),
		slider,
		sliderValue,
	)

	AddTab(tabs, tabName, container.NewScroll(content))
}

// CreateProgressTab создает вкладку с прогресс-барами
func CreateProgressTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Прогресс #%d", tabCounter)

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0.3)

	progressBarInfinite := widget.NewProgressBarInfinite()

	progressLabel := widget.NewLabel("Прогресс: 30%")

	startBtn := widget.NewButton("Запустить", func() {
		progressBarInfinite.Start()
		fmt.Println("[INFO] Запущен бесконечный прогресс")
	})

	stopBtn := widget.NewButton("Остановить", func() {
		progressBarInfinite.Stop()
		fmt.Println("[INFO] Остановлен бесконечный прогресс")
	})

	updateBtn := widget.NewButton("Обновить прогресс", func() {
		value := progressBar.Value + 0.1
		if value > 1.0 {
			value = 0.0
		}
		progressBar.SetValue(value)
		progressLabel.SetText(fmt.Sprintf("Прогресс: %.0f%%", value*100))
		fmt.Printf("[INFO] Прогресс обновлен: %.0f%%\n", value*100)
	})

	content := container.NewVBox(
		widget.NewLabel("Обычный прогресс-бар:"),
		progressBar,
		progressLabel,
		widget.NewSeparator(),
		widget.NewLabel("Бесконечный прогресс-бар:"),
		progressBarInfinite,
		container.NewHBox(startBtn, stopBtn),
		widget.NewSeparator(),
		updateBtn,
	)

	AddTab(tabs, tabName, container.NewScroll(content))
}

// CreateCustomTab создает кастомную вкладку
func CreateCustomTab(tabs *container.AppTabs) {
	tabCounter++
	tabName := fmt.Sprintf("Кастомная #%d", tabCounter)

	textArea := widget.NewMultiLineEntry()
	textArea.SetText("Это кастомная вкладка с различными виджетами.\n\nВы можете редактировать этот текст.")
	textArea.Wrapping = fyne.TextWrapWord

	clearBtn := widget.NewButton("Очистить", func() {
		textArea.SetText("")
		fmt.Println("[INFO] Текст очищен")
	})

	resetBtn := widget.NewButton("Сбросить", func() {
		textArea.SetText("Это кастомная вкладка с различными виджетами.\n\nВы можете редактировать этот текст.")
		fmt.Println("[INFO] Текст сброшен")
	})

	selectOptions := []string{"Опция 1", "Опция 2", "Опция 3", "Опция 4"}
	selectWidget := widget.NewSelect(selectOptions, func(selected string) {
		textArea.SetText(fmt.Sprintf("Выбрана опция: %s\n\nТекущий текст:\n%s", selected, textArea.Text))
		fmt.Printf("[INFO] Выбрана опция: %s\n", selected)
	})

	switchWidget := widget.NewCheck("Включить режим редактирования", func(checked bool) {
		if !checked {
			textArea.Disable()
		} else {
			textArea.Enable()
		}
		fmt.Printf("[INFO] Режим редактирования: %v\n", checked)
	})
	switchWidget.SetChecked(true)

	topContainer := container.NewVBox(
		widget.NewLabel("Выберите опцию:"),
		selectWidget,
		switchWidget,
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

	AddTab(tabs, tabName, content)
}

