package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tabManager struct {
	widget.BaseWidget
	tabs        []*TabData
	selectedTab int
	addCallback func()
}

type TabData struct {
	title   string
	content fyne.CanvasObject
}

func newTabManager() *tabManager {
	tm := &tabManager{
		tabs:        make([]*TabData, 0),
		selectedTab: -1,
	}
	tm.ExtendBaseWidget(tm)
	return tm
}

func (tm *tabManager) AddTab(title string, content fyne.CanvasObject) {
	tabCounter++
	if title == "" {
		title = "Tab"
	}

	tab := &TabData{
		title:   title,
		content: content,
	}
	tm.tabs = append(tm.tabs, tab)
	tm.selectedTab = len(tm.tabs) - 1
	tm.Refresh()
	log.Printf("Tab added: %s", title)
}

func (tm *tabManager) RemoveTab(index int) {
	if index < 0 || index >= len(tm.tabs) {
		log.Printf("Error: attempt to remove non-existent tab: index %d, total tabs: %d", index, len(tm.tabs))
		return
	}

	title := tm.tabs[index].title

	tm.tabs = append(tm.tabs[:index], tm.tabs[index+1:]...)

	if len(tm.tabs) == 0 {
		tm.selectedTab = -1
	} else {
		if tm.selectedTab == index {
			if index > 0 {
				tm.selectedTab = index - 1
			} else {
				tm.selectedTab = 0
			}
		} else if tm.selectedTab > index {
			tm.selectedTab--
		}
	}

	log.Printf("Tab '%s' removed, remaining tabs: %d", title, len(tm.tabs))
	tm.Refresh()
}

func (tm *tabManager) SelectTab(index int) {
	if index >= 0 && index < len(tm.tabs) {
		tm.selectedTab = index
		tm.Refresh()
	}
}

func (tm *tabManager) SetAddButtonCallback(callback func()) {
	tm.addCallback = callback
}

func (tm *tabManager) UpdateTabContent(content fyne.CanvasObject) {
	if tm.selectedTab >= 0 && tm.selectedTab < len(tm.tabs) {
		tm.tabs[tm.selectedTab].content = content
		tm.Refresh()
	} else {
		log.Printf("Error: failed to update content: selectedTab=%d, len(tabs)=%d", tm.selectedTab, len(tm.tabs))
	}
}

func (tm *tabManager) UpdateTabTitle(title string) {
	if tm.selectedTab >= 0 && tm.selectedTab < len(tm.tabs) {
		tm.tabs[tm.selectedTab].title = title
		tm.Refresh()
	} else {
		log.Printf("Error: failed to update title: selectedTab=%d, len(tabs)=%d", tm.selectedTab, len(tm.tabs))
	}
}

func (tm *tabManager) CreateRenderer() fyne.WidgetRenderer {
	addButton := widget.NewButton("+", func() {
		if tm.addCallback != nil {
			tm.addCallback()
		}
	})
	addButton.Importance = widget.LowImportance

	return newTabManagerRenderer(tm, addButton)
}
