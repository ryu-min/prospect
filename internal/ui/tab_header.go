package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TabHeader - кастомный виджет заголовка таба с кнопкой закрытия внутри
type TabHeader struct {
	widget.BaseWidget
	title      string
	isSelected bool
	onSelect   func()
	onClose    func()
}

// NewTabHeader создает новый заголовок таба
func NewTabHeader(title string, isSelected bool, onSelect, onClose func()) *TabHeader {
	th := &TabHeader{
		title:      title,
		isSelected: isSelected,
		onSelect:   onSelect,
		onClose:    onClose,
	}
	th.ExtendBaseWidget(th)
	return th
}

// CreateRenderer создает рендерер для заголовка таба
func (th *TabHeader) CreateRenderer() fyne.WidgetRenderer {
	// Текст названия таба
	titleText := canvas.NewText(th.title, theme.ForegroundColor())
	if th.isSelected {
		titleText.TextStyle = fyne.TextStyle{Bold: true}
	}

	// Кнопка закрытия (стандартная кнопка с фоном)
	closeBtn := widget.NewButton("×", func() {
		if th.onClose != nil {
			th.onClose()
		}
	})
	closeBtn.Importance = widget.LowImportance

	// Создаем единый контейнер: текст слева, кнопка закрытия справа
	// Используем HBox чтобы они были рядом без разделения
	content := container.NewHBox(
		container.NewPadded(titleText), // текст с небольшими отступами
		closeBtn,                       // кнопка закрытия
	)

	// Создаем фон для таба (будет выделяться при выборе)
	background := canvas.NewRectangle(theme.ButtonColor())
	if th.isSelected {
		background.FillColor = theme.PrimaryColor()
	}

	// Создаем кнопку для выбора вкладки (на весь таб)
	selectBtn := widget.NewButton("", func() {
		if th.onSelect != nil {
			th.onSelect()
		}
	})
	selectBtn.Importance = widget.LowImportance
	if th.isSelected {
		selectBtn.Importance = widget.HighImportance
	}

	// Объединяем: фон, кнопка выбора и контент
	mainContent := container.NewStack(
		background, // фон снизу
		selectBtn,  // кнопка выбора (прозрачная, на весь таб)
		content,    // контент сверху (текст + кнопка закрытия)
	)

	return &tabHeaderRenderer{
		header:     mainContent,
		titleText:  titleText,
		closeBtn:   closeBtn,
		selectBtn:  selectBtn,
		background: background,
		tabHeader:  th,
	}
}

type tabHeaderRenderer struct {
	header     fyne.CanvasObject
	titleText  *canvas.Text
	closeBtn   *widget.Button
	selectBtn  *widget.Button
	background *canvas.Rectangle
	tabHeader  *TabHeader
}

func (r *tabHeaderRenderer) Layout(size fyne.Size) {
	r.header.Resize(size)
}

func (r *tabHeaderRenderer) MinSize() fyne.Size {
	return r.header.MinSize()
}

func (r *tabHeaderRenderer) Refresh() {
	// Обновляем текст
	r.titleText.Text = r.tabHeader.title

	// Обновляем стиль текста и фона
	if r.tabHeader.isSelected {
		r.titleText.TextStyle = fyne.TextStyle{Bold: true}
		r.titleText.Color = theme.BackgroundColor()
		r.background.FillColor = theme.PrimaryColor()
	} else {
		r.titleText.TextStyle = fyne.TextStyle{}
		r.titleText.Color = theme.ForegroundColor()
		r.background.FillColor = theme.ButtonColor()
	}

	// Обновляем важность кнопки выбора
	if r.tabHeader.isSelected {
		r.selectBtn.Importance = widget.HighImportance
	} else {
		r.selectBtn.Importance = widget.LowImportance
	}

	r.header.Refresh()
}

func (r *tabHeaderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.header}
}

func (r *tabHeaderRenderer) Destroy() {}
