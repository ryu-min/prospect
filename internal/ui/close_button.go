package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type closeButton struct {
	widget.BaseWidget
	onClose   func()
	isHovered bool
	tabHeader *tabHeader
}

func newCloseButton(onClose func(), tabHeader *tabHeader) *closeButton {
	cb := &closeButton{
		onClose:   onClose,
		tabHeader: tabHeader,
	}
	cb.ExtendBaseWidget(cb)
	return cb
}

func (cb *closeButton) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText("×", theme.ForegroundColor())
	text.Alignment = fyne.TextAlignCenter

	background := canvas.NewRectangle(color.Transparent)

	return &closeButtonRenderer{
		closeButton: cb,
		text:        text,
		background:  background,
		container:   container.NewStack(background, container.NewCenter(text)),
	}
}

type closeButtonRenderer struct {
	closeButton *closeButton
	text        *canvas.Text
	background  *canvas.Rectangle
	container   fyne.CanvasObject
}

func (r *closeButtonRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

func (r *closeButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(24, 24)
}

func (r *closeButtonRenderer) Refresh() {
	r.text.Refresh()
	r.background.Refresh()
}

func (r *closeButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *closeButtonRenderer) Destroy() {}

func (cb *closeButton) MouseIn(*fyne.PointEvent) {
	cb.isHovered = true
	cb.Refresh()
}

func (cb *closeButton) MouseOut() {
	cb.isHovered = false
	cb.Refresh()
}

func (cb *closeButton) MouseDown(*fyne.PointEvent) {}

func (cb *closeButton) MouseUp(*fyne.PointEvent) {}

func (cb *closeButton) Tapped(*fyne.PointEvent) {
	if cb.onClose != nil {
		cb.onClose()
	}
}
