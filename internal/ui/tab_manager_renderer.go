package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type tabManagerRenderer struct {
	tabs        *tabManager
	header      fyne.CanvasObject
	contentArea fyne.CanvasObject
	mainContent fyne.CanvasObject
	addButton   *widget.Button
}

func newTabManagerRenderer(tm *tabManager, addButton *widget.Button) *tabManagerRenderer {
	return &tabManagerRenderer{
		tabs:      tm,
		addButton: addButton,
	}
}

func (r *tabManagerRenderer) Layout(size fyne.Size) {
	if r.mainContent != nil {
		r.mainContent.Resize(size)
	}
}

func (r *tabManagerRenderer) MinSize() fyne.Size {
	if r.mainContent != nil {
		return r.mainContent.MinSize()
	}
	return fyne.NewSize(100, 100)
}

func (r *tabManagerRenderer) Refresh() {
	tabButtons := make([]fyne.CanvasObject, 0)
	for i, tab := range r.tabs.tabs {
		tabIndex := i
		isSelected := r.tabs.selectedTab == tabIndex

		tabHeader := newTabHeader(
			tab.title,
			isSelected,
			func() {
				if tabIndex >= 0 && tabIndex < len(r.tabs.tabs) {
					r.tabs.SelectTab(tabIndex)
				}
			},
			func() {
				if tabIndex >= 0 && tabIndex < len(r.tabs.tabs) {
					r.tabs.RemoveTab(tabIndex)
				}
			},
		)

		tabButtons = append(tabButtons, tabHeader)
	}

	headerContent := make([]fyne.CanvasObject, 0)
	if len(tabButtons) > 0 {
		headerContent = append(headerContent, tabButtons...)
	}
	headerContent = append(headerContent, r.addButton)
	r.header = container.NewHBox(headerContent...)

	if r.tabs.selectedTab >= 0 && r.tabs.selectedTab < len(r.tabs.tabs) {
		r.contentArea = r.tabs.tabs[r.tabs.selectedTab].content
		if r.contentArea != nil {
			r.contentArea.Refresh()
		}
	} else {
		r.contentArea = widget.NewLabel("No open tabs")
	}

	r.mainContent = container.NewBorder(
		r.header,
		nil,
		nil,
		nil,
		r.contentArea,
	)
}

func (r *tabManagerRenderer) Objects() []fyne.CanvasObject {
	if r.mainContent != nil {
		return []fyne.CanvasObject{r.mainContent}
	}
	return []fyne.CanvasObject{}
}

func (r *tabManagerRenderer) Destroy() {}
