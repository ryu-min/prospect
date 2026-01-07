package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type tabHeader struct {
	widget.BaseWidget
	title      string
	isSelected bool
	onSelect   func()
	onClose    func()
}

func newTabHeader(title string, isSelected bool, onSelect, onClose func()) *tabHeader {
	th := &tabHeader{
		title:      title,
		isSelected: isSelected,
		onSelect:   onSelect,
		onClose:    onClose,
	}
	th.ExtendBaseWidget(th)
	return th
}

func (th *tabHeader) CreateRenderer() fyne.WidgetRenderer {
	titleText := canvas.NewText(th.title, theme.ForegroundColor())
	if th.isSelected {
		titleText.TextStyle = fyne.TextStyle{Bold: true}
	}

	closeBtn := newCloseButton(func() {
		if th.onClose != nil {
			th.onClose()
		}
	}, th)

	content := container.NewHBox(
		container.NewPadded(titleText),
		closeBtn,
	)

	background := canvas.NewRectangle(theme.ButtonColor())
	if th.isSelected {
		background.FillColor = theme.PrimaryColor()
	}

	selectBtn := widget.NewButton("", func() {
		if th.onSelect != nil {
			th.onSelect()
		}
	})
	selectBtn.Importance = widget.LowImportance
	if th.isSelected {
		selectBtn.Importance = widget.HighImportance
	}

	mainContent := container.NewStack(
		background,
		selectBtn,
		content,
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
	closeBtn   *closeButton
	selectBtn  *widget.Button
	background *canvas.Rectangle
	tabHeader  *tabHeader
}

func (r *tabHeaderRenderer) Layout(size fyne.Size) {
	r.header.Resize(size)
}

func (r *tabHeaderRenderer) MinSize() fyne.Size {
	return r.header.MinSize()
}

func (r *tabHeaderRenderer) Refresh() {
	r.titleText.Text = r.tabHeader.title

	if r.tabHeader.isSelected {
		r.titleText.TextStyle = fyne.TextStyle{Bold: true}
		r.titleText.Color = theme.BackgroundColor()
		r.background.FillColor = theme.PrimaryColor()
	} else {
		r.titleText.TextStyle = fyne.TextStyle{}
		r.titleText.Color = theme.ForegroundColor()
		r.background.FillColor = theme.ButtonColor()
	}

	if r.tabHeader.isSelected {
		r.selectBtn.Importance = widget.HighImportance
	} else {
		r.selectBtn.Importance = widget.LowImportance
	}

	r.closeBtn.Refresh()

	r.header.Refresh()
}

func (r *tabHeaderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.header}
}

func (r *tabHeaderRenderer) Destroy() {}
