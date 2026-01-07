package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const (
	nameColumnWidth = 120
	typeColumnWidth = 100
	columnSpacing   = 10
)

type protoFieldEditor struct {
	widget.BaseWidget
	nameLabel      *widget.Label
	typeCombo      *widget.Select
	entry          *widget.Entry
	uid            widget.TreeNodeID
	adapter        *protoTreeAdapter
	availableTypes []string
}

func newProtoFieldEditor(uid widget.TreeNodeID, adapter *protoTreeAdapter) *protoFieldEditor {
	availableTypes := []string{"string", "number", "bool"}
	ew := &protoFieldEditor{
		uid:            uid,
		adapter:        adapter,
		nameLabel:      widget.NewLabel(""),
		typeCombo:      widget.NewSelect(availableTypes, nil),
		entry:          widget.NewEntry(),
		availableTypes: availableTypes,
	}
	ew.entry.OnChanged = func(value string) {
		adapter.updateNodeValue(uid, value, "")
	}
	ew.ExtendBaseWidget(ew)
	return ew
}

func (ew *protoFieldEditor) CreateRenderer() fyne.WidgetRenderer {
	return &protoFieldEditorRenderer{
		widget:    ew,
		nameLabel: ew.nameLabel,
		typeCombo: ew.typeCombo,
		entry:     ew.entry,
		objects:   []fyne.CanvasObject{ew.nameLabel, ew.typeCombo, ew.entry},
	}
}

type protoFieldEditorRenderer struct {
	widget    *protoFieldEditor
	nameLabel *widget.Label
	typeCombo *widget.Select
	entry     *widget.Entry
	objects   []fyne.CanvasObject
}

func (r *protoFieldEditorRenderer) Layout(size fyne.Size) {
	namePos := fyne.NewPos(0, (size.Height-r.nameLabel.MinSize().Height)/2)
	r.nameLabel.Move(namePos)
	r.nameLabel.Resize(fyne.NewSize(float32(nameColumnWidth), r.nameLabel.MinSize().Height))

	typePos := fyne.NewPos(float32(nameColumnWidth+columnSpacing), (size.Height-r.typeCombo.MinSize().Height)/2)
	r.typeCombo.Move(typePos)
	r.typeCombo.Resize(fyne.NewSize(float32(typeColumnWidth), r.typeCombo.MinSize().Height))

	entryX := float32(nameColumnWidth + typeColumnWidth + columnSpacing*2)
	entryWidth := size.Width - entryX
	entryPos := fyne.NewPos(entryX, (size.Height-r.entry.MinSize().Height)/2)
	r.entry.Move(entryPos)
	r.entry.Resize(fyne.NewSize(entryWidth, r.entry.MinSize().Height))
}

func (r *protoFieldEditorRenderer) MinSize() fyne.Size {
	nameSize := r.nameLabel.MinSize()
	typeSize := r.typeCombo.MinSize()
	entrySize := r.entry.MinSize()

	width := float32(nameColumnWidth + typeColumnWidth + int(entrySize.Width) + columnSpacing*2)
	height := fyne.Max(fyne.Max(nameSize.Height, typeSize.Height), entrySize.Height)

	return fyne.NewSize(width, height)
}

func (r *protoFieldEditorRenderer) Refresh() {
	r.nameLabel.Refresh()
	r.typeCombo.Refresh()
	r.entry.Refresh()
}

func (r *protoFieldEditorRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *protoFieldEditorRenderer) Destroy() {}

