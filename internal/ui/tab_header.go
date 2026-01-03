package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CloseButton - кастомная кнопка закрытия с почти прозрачным фоном при наведении
type CloseButton struct {
	widget.BaseWidget
	onClose   func()
	isHovered bool
	tabHeader *TabHeader // ссылка на родительский таб для определения цвета текста
}

// NewCloseButton создает новую кнопку закрытия
func NewCloseButton(onClose func(), tabHeader *TabHeader) *CloseButton {
	cb := &CloseButton{
		onClose:   onClose,
		tabHeader: tabHeader,
	}
	cb.ExtendBaseWidget(cb)
	return cb
}

// CreateRenderer создает рендерер для кнопки закрытия
func (cb *CloseButton) CreateRenderer() fyne.WidgetRenderer {
	// Текст "×"
	text := canvas.NewText("×", theme.ForegroundColor())
	text.Alignment = fyne.TextAlignCenter

	// Фон кнопки
	background := canvas.NewRectangle(color.Transparent)

	return &closeButtonRenderer{
		closeButton: cb,
		text:        text,
		background:  background,
		container:   container.NewStack(background, container.NewCenter(text)),
	}
}

type closeButtonRenderer struct {
	closeButton *CloseButton
	text        *canvas.Text
	background  *canvas.Rectangle
	container   fyne.CanvasObject
}

func (r *closeButtonRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

func (r *closeButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(24, 24) // Минимальный размер кнопки
}

func (r *closeButtonRenderer) Refresh() {
	// Определяем цвет текста в зависимости от состояния таба
	// if r.closeButton.tabHeader != nil && r.closeButton.tabHeader.isSelected {
	// 	r.text.Color = theme.BackgroundColor()
	// } else {
	// 	r.text.Color = theme.ForegroundColor()
	// }

	// Обновляем фон при наведении - почти прозрачный
	// if r.closeButton.isHovered {
	// 	// Полупрозрачный фон при наведении (alpha ~0.25)
	// 	// Конвертируем цвет темы в RGBA
	// 	baseColor := theme.ButtonColor()
	// 	cr, cg, cb, _ := baseColor.RGBA()
	// 	r.background.FillColor = color.RGBA{
	// 		R: uint8(cr),
	// 		G: uint8(cg),
	// 		B: uint8(cb),
	// 		A: 150,
	// 	}
	// } else {
	// 	r.background.FillColor = color.Transparent
	// }

	r.text.Refresh()
	r.background.Refresh()
}

func (r *closeButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *closeButtonRenderer) Destroy() {}

// MouseIn обрабатывает наведение мыши
func (cb *CloseButton) MouseIn(*fyne.PointEvent) {
	cb.isHovered = true
	cb.Refresh()
}

// MouseOut обрабатывает уход мыши
func (cb *CloseButton) MouseOut() {
	cb.isHovered = false
	cb.Refresh()
}

// MouseDown обрабатывает нажатие мыши
func (cb *CloseButton) MouseDown(*fyne.PointEvent) {}

// MouseUp обрабатывает отпускание мыши
func (cb *CloseButton) MouseUp(*fyne.PointEvent) {}

// Tapped обрабатывает клик
func (cb *CloseButton) Tapped(*fyne.PointEvent) {
	if cb.onClose != nil {
		cb.onClose()
	}
}

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

	// Кнопка закрытия (кастомная с почти прозрачным фоном при наведении)
	closeBtn := NewCloseButton(func() {
		if th.onClose != nil {
			th.onClose()
		}
	}, th)

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
	closeBtn   *CloseButton
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

	// Обновляем кнопку закрытия (чтобы обновился цвет текста)
	r.closeBtn.Refresh()

	r.header.Refresh()
}

func (r *tabHeaderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.header}
}

func (r *tabHeaderRenderer) Destroy() {}
