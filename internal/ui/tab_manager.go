package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type tabManager struct {
	widget.BaseWidget
	tabs         []*tabData
	selectedTab  int
	addCallback  func()
	toolbarMgr   *toolbarManager
}

type tabData struct {
	title           string
	content         fyne.CanvasObject
	filePath        string
	toolbarCallbacks *toolbarCallbacks
}

type toolbarCallbacks struct {
	openCallback         func()
	saveCallback         func()
	applySchemaCallback  func()
	exportJSONCallback   func()
	exportSchemaCallback func()
}

func newTabManager() *tabManager {
	tm := &tabManager{
		tabs:        make([]*tabData, 0),
		selectedTab: -1,
		toolbarMgr:  newToolbarManager(),
	}
	tm.ExtendBaseWidget(tm)
	return tm
}

func (tm *tabManager) AddTab(title string, content fyne.CanvasObject) {
	tm.addTabWithoutSave(title, content)
	go saveTabState(tm)
}

func (tm *tabManager) addTabWithoutSave(title string, content fyne.CanvasObject) {
	tm.addTabWithPathWithoutSave(title, content, "")
}

func (tm *tabManager) addTabWithPathWithoutSave(title string, content fyne.CanvasObject, filePath string) {
	tabCounter++
	if title == "" {
		title = "Tab"
	}

	tab := &tabData{
		title:    title,
		content:  content,
		filePath: filePath,
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
	go saveTabState(tm)
}

func (tm *tabManager) SelectTab(index int) {
	tm.selectTabWithoutSave(index)
	go saveTabState(tm)
}

func (tm *tabManager) selectTabWithoutSave(index int) {
	if index >= 0 && index < len(tm.tabs) {
		tm.selectedTab = index
		if tm.tabs[index].toolbarCallbacks != nil {
			callbacks := tm.tabs[index].toolbarCallbacks
			if callbacks.openCallback != nil {
				tm.toolbarMgr.SetOpenCallback(callbacks.openCallback)
			}
			if callbacks.saveCallback != nil {
				tm.toolbarMgr.SetSaveCallback(callbacks.saveCallback)
			}
			if callbacks.applySchemaCallback != nil {
				tm.toolbarMgr.SetApplySchemaCallback(callbacks.applySchemaCallback)
			}
			if callbacks.exportJSONCallback != nil {
				tm.toolbarMgr.SetExportJSONCallback(callbacks.exportJSONCallback)
			}
			if callbacks.exportSchemaCallback != nil {
				tm.toolbarMgr.SetExportSchemaCallback(callbacks.exportSchemaCallback)
			}
		}
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
		go saveTabState(tm)
	} else {
		log.Printf("Error: failed to update title: selectedTab=%d, len(tabs)=%d", tm.selectedTab, len(tm.tabs))
	}
}

func (tm *tabManager) SetTabFilePath(filePath string) {
	if tm.selectedTab >= 0 && tm.selectedTab < len(tm.tabs) {
		tm.tabs[tm.selectedTab].filePath = filePath
		go saveTabState(tm)
	}
}

func (tm *tabManager) GetToolbarManager() *toolbarManager {
	return tm.toolbarMgr
}

func (tm *tabManager) SetCurrentTabToolbarCallbacks(callbacks *toolbarCallbacks) {
	if tm.selectedTab >= 0 && tm.selectedTab < len(tm.tabs) {
		tm.tabs[tm.selectedTab].toolbarCallbacks = callbacks
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
