package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BrowserTabs - кастомная система табов как в браузере
type BrowserTabs struct {
	widget.BaseWidget
	tabs        []*TabData
	selectedTab int
	addCallback func()
}

// TabData - данные о вкладке
type TabData struct {
	title   string
	content fyne.CanvasObject
}

// NewBrowserTabs создает новую систему табов
func NewBrowserTabs() *BrowserTabs {
	bt := &BrowserTabs{
		tabs:        make([]*TabData, 0),
		selectedTab: -1,
	}
	bt.ExtendBaseWidget(bt)
	return bt
}

// AddTab добавляет новую вкладку
func (bt *BrowserTabs) AddTab(title string, content fyne.CanvasObject) {
	tabCounter++
	if title == "" {
		title = "Tab"
	}

	tab := &TabData{
		title:   title,
		content: content,
	}
	bt.tabs = append(bt.tabs, tab)
	bt.selectedTab = len(bt.tabs) - 1
	bt.Refresh()
	log.Printf("Tab added: %s", title)
}

// RemoveTab удаляет вкладку по индексу
func (bt *BrowserTabs) RemoveTab(index int) {
	if index < 0 || index >= len(bt.tabs) {
		log.Printf("Error: attempt to remove non-existent tab: index %d, total tabs: %d", index, len(bt.tabs))
		return
	}

	title := bt.tabs[index].title

	// Удаляем вкладку
	bt.tabs = append(bt.tabs[:index], bt.tabs[index+1:]...)

	// Обновляем выбранную вкладку
	if len(bt.tabs) == 0 {
		// Нет вкладок
		bt.selectedTab = -1
	} else {
		// Есть вкладки - корректируем индекс выбранной
		if bt.selectedTab == index {
			// Удалена выбранная вкладка - выбираем предыдущую или последнюю
			if index > 0 {
				bt.selectedTab = index - 1
			} else {
				bt.selectedTab = 0
			}
		} else if bt.selectedTab > index {
			// Удалена вкладка перед выбранной - уменьшаем индекс
			bt.selectedTab--
		}
		// Если bt.selectedTab < index, ничего не меняем
	}

	log.Printf("Tab '%s' removed, remaining tabs: %d", title, len(bt.tabs))
	bt.Refresh()
}

// SelectTab выбирает вкладку по индексу
func (bt *BrowserTabs) SelectTab(index int) {
	if index >= 0 && index < len(bt.tabs) {
		bt.selectedTab = index
		bt.Refresh()
	}
}

// SetAddButtonCallback устанавливает обработчик для кнопки "+"
func (bt *BrowserTabs) SetAddButtonCallback(callback func()) {
	bt.addCallback = callback
}

// UpdateTabContent обновляет контент выбранной вкладки
func (bt *BrowserTabs) UpdateTabContent(content fyne.CanvasObject) {
	if bt.selectedTab >= 0 && bt.selectedTab < len(bt.tabs) {
		bt.tabs[bt.selectedTab].content = content
		bt.Refresh()
	} else {
		log.Printf("Error: failed to update content: selectedTab=%d, len(tabs)=%d", bt.selectedTab, len(bt.tabs))
	}
}

// UpdateTabTitle обновляет название выбранной вкладки
func (bt *BrowserTabs) UpdateTabTitle(title string) {
	if bt.selectedTab >= 0 && bt.selectedTab < len(bt.tabs) {
		bt.tabs[bt.selectedTab].title = title
		bt.Refresh()
	} else {
		log.Printf("Error: failed to update title: selectedTab=%d, len(tabs)=%d", bt.selectedTab, len(bt.tabs))
	}
}

// CreateRenderer создает рендерер
func (bt *BrowserTabs) CreateRenderer() fyne.WidgetRenderer {
	// Создаем кнопку "+" один раз при создании рендерера
	addButton := widget.NewButton("+", func() {
		if bt.addCallback != nil {
			bt.addCallback()
		}
	})
	addButton.Importance = widget.LowImportance

	return &browserTabsRenderer{
		tabs:      bt,
		addButton: addButton,
	}
}

type browserTabsRenderer struct {
	tabs        *BrowserTabs
	header      fyne.CanvasObject
	contentArea fyne.CanvasObject
	mainContent fyne.CanvasObject
	addButton   *widget.Button
}

func (r *browserTabsRenderer) Layout(size fyne.Size) {
	if r.mainContent != nil {
		r.mainContent.Resize(size)
	}
}

func (r *browserTabsRenderer) MinSize() fyne.Size {
	if r.mainContent != nil {
		return r.mainContent.MinSize()
	}
	return fyne.NewSize(100, 100)
}

func (r *browserTabsRenderer) Refresh() {
	// Создаем заголовки табов
	tabButtons := make([]fyne.CanvasObject, 0)
	for i, tab := range r.tabs.tabs {
		// Сохраняем индекс в локальную переменную для замыкания
		tabIndex := i
		isSelected := r.tabs.selectedTab == tabIndex

		// Создаем кастомный заголовок таба с кнопкой закрытия внутри
		tabHeader := NewTabHeader(
			tab.title,
			isSelected,
			func() {
				// Проверяем, что индекс все еще валиден
				if tabIndex >= 0 && tabIndex < len(r.tabs.tabs) {
					r.tabs.SelectTab(tabIndex)
				}
			},
			func() {
				// Проверяем, что индекс все еще валиден перед удалением
				if tabIndex >= 0 && tabIndex < len(r.tabs.tabs) {
					r.tabs.RemoveTab(tabIndex)
				}
			},
		)

		tabButtons = append(tabButtons, tabHeader)
	}

	// Создаем заголовок с табами и кнопкой "+" справа
	// Кнопка "+" всегда должна быть видна и работать, даже если нет вкладок
	headerContent := make([]fyne.CanvasObject, 0)
	if len(tabButtons) > 0 {
		headerContent = append(headerContent, tabButtons...)
	}
	// Кнопка "+" всегда добавляется
	headerContent = append(headerContent, r.addButton)
	r.header = container.NewHBox(headerContent...)

	// Создаем область содержимого
	if r.tabs.selectedTab >= 0 && r.tabs.selectedTab < len(r.tabs.tabs) {
		r.contentArea = r.tabs.tabs[r.tabs.selectedTab].content
		// ВАЖНО: Обновляем контент после установки
		if r.contentArea != nil {
			r.contentArea.Refresh()
		}
	} else {
		r.contentArea = widget.NewLabel("No open tabs")
	}

	// Основной контейнер: заголовок сверху, содержимое снизу
	r.mainContent = container.NewBorder(
		r.header,      // верх - заголовки табов и кнопка "+"
		nil,           // низ
		nil,           // лево
		nil,           // право
		r.contentArea, // центр - содержимое выбранной вкладки
	)

}

func (r *browserTabsRenderer) Objects() []fyne.CanvasObject {
	if r.mainContent != nil {
		return []fyne.CanvasObject{r.mainContent}
	}
	return []fyne.CanvasObject{}
}

func (r *browserTabsRenderer) Destroy() {}
