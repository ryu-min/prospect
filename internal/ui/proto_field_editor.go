package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

const (
	nameColumnWidth = 150
	typeColumnWidth = 120
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
	showEntry      bool
}

func newProtoFieldEditor(uid widget.TreeNodeID, adapter *protoTreeAdapter, messageTypes []string) *protoFieldEditor {
	availableTypes := []string{"string", "number", "bool"}
	availableTypes = append(availableTypes, messageTypes...)
	nameLabel := widget.NewLabel("")
	nameLabel.Wrapping = fyne.TextTruncate
	ew := &protoFieldEditor{
		uid:            uid,
		adapter:        adapter,
		nameLabel:      nameLabel,
		typeCombo:      widget.NewSelect(availableTypes, nil),
		entry:          widget.NewEntry(),
		availableTypes: availableTypes,
		showEntry:      true,
	}
	ew.entry.OnChanged = func(value string) {
		adapter.updateNodeValue(uid, value, "")
	}
	ew.ExtendBaseWidget(ew)
	return ew
}

func (ew *protoFieldEditor) SetEntryVisible(visible bool) {
	if ew.showEntry != visible {
		ew.showEntry = visible
		ew.Refresh()
	}
}

func (ew *protoFieldEditor) CreateRenderer() fyne.WidgetRenderer {
	return &protoFieldEditorRenderer{
		widget:    ew,
		nameLabel: ew.nameLabel,
		typeCombo: ew.typeCombo,
		entry:     ew.entry,
	}
}

type protoFieldEditorRenderer struct {
	widget    *protoFieldEditor
	nameLabel *widget.Label
	typeCombo *widget.Select
	entry     *widget.Entry
}

func (r *protoFieldEditorRenderer) Layout(size fyne.Size) {
	namePos := fyne.NewPos(0, (size.Height-r.nameLabel.MinSize().Height)/2)
	r.nameLabel.Move(namePos)
	r.nameLabel.Resize(fyne.NewSize(float32(nameColumnWidth), r.nameLabel.MinSize().Height))

	typePos := fyne.NewPos(float32(nameColumnWidth+columnSpacing), (size.Height-r.typeCombo.MinSize().Height)/2)
	r.typeCombo.Move(typePos)
	r.typeCombo.Resize(fyne.NewSize(float32(typeColumnWidth), r.typeCombo.MinSize().Height))

	if r.widget.showEntry {
		entryX := float32(nameColumnWidth + typeColumnWidth + columnSpacing*2)
		entryWidth := size.Width - entryX
		entryPos := fyne.NewPos(entryX, (size.Height-r.entry.MinSize().Height)/2)
		r.entry.Move(entryPos)
		r.entry.Resize(fyne.NewSize(entryWidth, r.entry.MinSize().Height))
	}
}

func (r *protoFieldEditorRenderer) MinSize() fyne.Size {
	nameSize := r.nameLabel.MinSize()
	typeSize := r.typeCombo.MinSize()

	width := float32(nameColumnWidth + typeColumnWidth + columnSpacing)
	height := fyne.Max(nameSize.Height, typeSize.Height)

	if r.widget.showEntry {
		entrySize := r.entry.MinSize()
		width += float32(int(entrySize.Width) + columnSpacing)
		height = fyne.Max(height, entrySize.Height)
	}

	return fyne.NewSize(width, height)
}

func (r *protoFieldEditorRenderer) Refresh() {
	r.nameLabel.Refresh()
	r.typeCombo.Refresh()
	r.entry.Refresh()
}

func (r *protoFieldEditorRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.nameLabel, r.typeCombo}
	if r.widget.showEntry {
		objects = append(objects, r.entry)
	}
	return objects
}

func (r *protoFieldEditorRenderer) Destroy() {}

