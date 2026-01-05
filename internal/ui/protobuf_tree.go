package ui

import (
	"fmt"
	"strconv"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Global flag for "don't ask again" for type change confirmation
var dontAskTypeChangeConfirmation bool

// ProtobufTreeAdapter адаптирует TreeNode для widget.Tree
type ProtobufTreeAdapter struct {
	tree        *protobuf.TreeNode
	editWidgets map[widget.TreeNodeID]*EditableNodeWidget
	window      fyne.Window // Window for dialogs
}

// EditableNodeWidget - виджет для редактирования значения узла
type EditableNodeWidget struct {
	widget.BaseWidget
	nameLabel      *widget.Label  // Имя поля
	typeCombo      *widget.Select // Тип поля (ComboBox)
	entry          *widget.Entry  // Значение для редактирования
	uid            widget.TreeNodeID
	adapter        *ProtobufTreeAdapter
	availableTypes []string // Доступные типы
}

// Константы для ширины колонок
const (
	nameColumnWidth = 120 // Ширина колонки имени
	typeColumnWidth = 100 // Ширина колонки типа (увеличено для "number")
	columnSpacing   = 10  // Отступ между колонками
)

// NewEditableNodeWidget создает новый редактируемый виджет узла
func NewEditableNodeWidget(uid widget.TreeNodeID, adapter *ProtobufTreeAdapter) *EditableNodeWidget {
	availableTypes := []string{"string", "number", "bool"}
	ew := &EditableNodeWidget{
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

// CreateRenderer создает рендерер для редактируемого виджета
func (ew *EditableNodeWidget) CreateRenderer() fyne.WidgetRenderer {
	return &editableNodeRenderer{
		widget:    ew,
		nameLabel: ew.nameLabel,
		typeCombo: ew.typeCombo,
		entry:     ew.entry,
		objects:   []fyne.CanvasObject{ew.nameLabel, ew.typeCombo, ew.entry},
	}
}

type editableNodeRenderer struct {
	widget    *EditableNodeWidget
	nameLabel *widget.Label
	typeCombo *widget.Select
	entry     *widget.Entry
	objects   []fyne.CanvasObject
}

func (r *editableNodeRenderer) Layout(size fyne.Size) {
	// Все колонки начинаются на одном уровне (с учетом отступов дерева)
	// Колонка имени
	namePos := fyne.NewPos(0, (size.Height-r.nameLabel.MinSize().Height)/2)
	r.nameLabel.Move(namePos)
	r.nameLabel.Resize(fyne.NewSize(float32(nameColumnWidth), r.nameLabel.MinSize().Height))

	// Колонка типа (ComboBox)
	typePos := fyne.NewPos(float32(nameColumnWidth+columnSpacing), (size.Height-r.typeCombo.MinSize().Height)/2)
	r.typeCombo.Move(typePos)
	r.typeCombo.Resize(fyne.NewSize(float32(typeColumnWidth), r.typeCombo.MinSize().Height))

	// Колонка значения (Entry) - занимает оставшееся пространство
	entryX := float32(nameColumnWidth + typeColumnWidth + columnSpacing*2)
	entryWidth := size.Width - entryX
	entryPos := fyne.NewPos(entryX, (size.Height-r.entry.MinSize().Height)/2)
	r.entry.Move(entryPos)
	r.entry.Resize(fyne.NewSize(entryWidth, r.entry.MinSize().Height))
}

func (r *editableNodeRenderer) MinSize() fyne.Size {
	nameSize := r.nameLabel.MinSize()
	typeSize := r.typeCombo.MinSize()
	entrySize := r.entry.MinSize()

	width := float32(nameColumnWidth + typeColumnWidth + int(entrySize.Width) + columnSpacing*2)
	height := fyne.Max(fyne.Max(nameSize.Height, typeSize.Height), entrySize.Height)

	return fyne.NewSize(width, height)
}

func (r *editableNodeRenderer) Refresh() {
	r.nameLabel.Refresh()
	r.typeCombo.Refresh()
	r.entry.Refresh()
}

func (r *editableNodeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *editableNodeRenderer) Destroy() {}

// NewProtobufTreeAdapter создает новый адаптер для дерева
func NewProtobufTreeAdapter(tree *protobuf.TreeNode) *ProtobufTreeAdapter {
	return &ProtobufTreeAdapter{
		tree:        tree,
		editWidgets: make(map[widget.TreeNodeID]*EditableNodeWidget),
		window:      nil,
	}
}

// SetWindow устанавливает окно для диалогов
func (a *ProtobufTreeAdapter) SetWindow(window fyne.Window) {
	a.window = window
}

// ChildUIDs возвращает UID дочерних узлов
func (a *ProtobufTreeAdapter) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	// В Fyne widget.Tree использует пустую строку "" для root
	// Преобразуем пустую строку в "root" для нашего дерева
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		return nil
	}

	children := make([]widget.TreeNodeID, 0, len(node.Children))
	for i := range node.Children {
		// Для root используем пустую строку + индекс, для остальных - обычный формат
		if actualUID == "root" {
			childUID := fmt.Sprintf("%d", i)
			children = append(children, childUID)
		} else {
			childUID := fmt.Sprintf("%s:%d", uid, i)
			children = append(children, childUID)
		}
	}
	return children
}

// IsBranch возвращает true, если узел является веткой (имеет детей)
func (a *ProtobufTreeAdapter) IsBranch(uid widget.TreeNodeID) bool {
	// В Fyne widget.Tree использует пустую строку "" для root
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		return false
	}
	return len(node.Children) > 0
}

// CreateNode создает виджет для узла
func (a *ProtobufTreeAdapter) CreateNode(branch bool) fyne.CanvasObject {
	// Для веток (сообщений) используем просто Label
	if branch {
		label := widget.NewLabel("")
		label.Wrapping = fyne.TextWrapWord
		return label
	}

	// Для листьев создаем редактируемый виджет
	// UID будет установлен в UpdateNode
	ew := NewEditableNodeWidget("", a)
	return ew
}

// UpdateNode обновляет виджет узла
func (a *ProtobufTreeAdapter) UpdateNode(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
	// В Fyne widget.Tree использует пустую строку "" для root
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		if label, ok := obj.(*widget.Label); ok {
			label.SetText(fmt.Sprintf("ERROR: Node not found ('%s')", uid))
		}
		return
	}

	// Для веток (сообщений) используем Label
	if branch || node.Name == "root" {
		if label, ok := obj.(*widget.Label); ok {
			text := ""
			if node.Name == "root" {
				text = "Protobuf Root"
				if len(node.Children) == 0 {
					text = "Protobuf Root (no data)"
				}
			} else {
				text = fmt.Sprintf("%s (field_%d, %s)", node.Name, node.FieldNum, node.Type)
				if node.IsRepeated {
					text += " [repeated]"
				}
				if len(node.Children) > 0 {
					text += fmt.Sprintf(" [%d children]", len(node.Children))
				}
			}
			label.SetText(text)
		}
		return
	}

	// Для листьев (примитивных значений) используем EditableNodeWidget
	if editWidget, ok := obj.(*EditableNodeWidget); ok {
		editWidget.uid = actualUID
		a.editWidgets[actualUID] = editWidget

		// Устанавливаем имя поля в первую колонку
		nameText := fmt.Sprintf("field_%d", node.FieldNum)
		editWidget.nameLabel.SetText(nameText)

		// Устанавливаем тип поля во вторую колонку (ComboBox)
		typeText := node.Type
		// Проверяем, что тип есть в списке доступных
		typeExists := false
		for _, t := range editWidget.availableTypes {
			if t == typeText {
				typeExists = true
				break
			}
		}
		// Если типа нет в списке, добавляем его (для неизвестных типов)
		if !typeExists && typeText != "" {
			editWidget.availableTypes = append(editWidget.availableTypes, typeText)
			editWidget.typeCombo.Options = editWidget.availableTypes
		}

		// Временно отключаем OnChanged для ComboBox, чтобы не триггерить при установке значения
		editWidget.typeCombo.OnChanged = nil
		editWidget.typeCombo.SetSelected(typeText)

		// Устанавливаем обработчик изменения типа
		editWidget.typeCombo.OnChanged = func(selectedType string) {
			if selectedType == "" || selectedType == node.Type {
				return // Не изменился или пустое значение
			}
			// Вызываем логику изменения типа
			a.handleTypeChange(actualUID, node.Type, selectedType)
		}

		// Получаем текущее значение из узла (источник истины)
		// ВАЖНО: Всегда читаем из узла, а не из виджета
		valueStr := a.nodeValueToString(node)

		// Временно отключаем OnChanged, чтобы не триггерить обновление при SetText
		editWidget.entry.OnChanged = nil
		editWidget.entry.SetText(valueStr)

		// Сохраняем последнее валидное значение для отката
		lastValidValue := valueStr
		var isUpdating bool // Флаг для предотвращения рекурсии

		// Устанавливаем обработчик с правильным типом после установки значения
		fieldType := node.Type
		editWidget.entry.OnChanged = func(value string) {
			// Предотвращаем рекурсию при откате
			if isUpdating {
				return
			}

			// Валидируем значение в зависимости от типа
			if !a.validateValue(value, fieldType) {
				// Если значение невалидно, откатываем к последнему валидному
				isUpdating = true
				editWidget.entry.SetText(lastValidValue)
				isUpdating = false
				return
			}

			// Значение валидно, обновляем
			lastValidValue = value
			a.updateNodeValue(actualUID, value, fieldType)
		}
		editWidget.Refresh()
		return
	}
}

// updateEntryValidation обновляет валидацию в поле ввода при изменении типа
func (a *ProtobufTreeAdapter) updateEntryValidation(uid widget.TreeNodeID, newType string) {
	editWidget, ok := a.editWidgets[uid]
	if !ok {
		return
	}

	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	// Получаем текущее значение из виджета (может быть пустым после очистки)
	currentValue := editWidget.entry.Text
	lastValidValue := currentValue
	var isUpdating bool

	// Обновляем обработчик с новым типом
	editWidget.entry.OnChanged = nil
	editWidget.entry.OnChanged = func(value string) {
		// Предотвращаем рекурсию при откате
		if isUpdating {
			return
		}

		// Валидируем значение в зависимости от нового типа
		if !a.validateValue(value, newType) {
			// Если значение невалидно, откатываем к последнему валидному
			isUpdating = true
			editWidget.entry.SetText(lastValidValue)
			isUpdating = false
			return
		}

		// Значение валидно, обновляем
		lastValidValue = value
		a.updateNodeValue(uid, value, newType)
	}
}

// nodeValueToString преобразует значение узла в строку для отображения
func (a *ProtobufTreeAdapter) nodeValueToString(node *protobuf.TreeNode) string {
	if node.Value == nil {
		return ""
	}

	switch v := node.Value.(type) {
	case string:
		return v
	case bool:
		// Отображаем bool как "0" или "1"
		if v {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// validateValue проверяет, является ли значение валидным для данного типа
// Разрешает промежуточные состояния ввода (например, "-" для отрицательных чисел)
func (a *ProtobufTreeAdapter) validateValue(value string, fieldType string) bool {
	if value == "" {
		// Пустое значение допустимо для всех типов
		return true
	}

	switch fieldType {
	case "string":
		// Для строки любое значение валидно
		return true
	case "number":
		// Для числа разрешаем промежуточные состояния ввода
		// "-" - начало отрицательного числа
		if value == "-" {
			return true
		}
		// "." - начало десятичного числа
		if value == "." {
			return true
		}
		// "-." - начало отрицательного десятичного числа
		if value == "-." {
			return true
		}
		// Проверяем, что это валидное число или промежуточное состояние
		// Разрешаем паттерны типа "123.", "-123.", ".5", "-.5"
		trimmed := strings.TrimSpace(value)
		if strings.HasSuffix(trimmed, ".") {
			// Убираем точку в конце для проверки
			trimmed = strings.TrimSuffix(trimmed, ".")
		}
		// Пробуем распарсить как int
		_, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			// Пробуем как float
			_, err = strconv.ParseFloat(trimmed, 64)
			if err != nil {
				// Если не получилось, проверяем промежуточные состояния
				// Разрешаем паттерны типа "-", "123.", "-123."
				if trimmed == "-" || strings.HasSuffix(value, ".") {
					return true
				}
				return false
			}
		}
		return true
	case "bool":
		// Для bool только 0 и 1
		trimmed := strings.TrimSpace(value)
		return trimmed == "0" || trimmed == "1" || trimmed == ""
	default:
		// Для неизвестных типов разрешаем любое значение
		return true
	}
}

// detectTypeChange определяет, изменился ли тип значения
// ВАЖНО: Эта функция используется только для определения изменения типа при вводе значения
// Для полей типа bool/number значения "0" и "1" остаются в своем типе
func (a *ProtobufTreeAdapter) detectTypeChange(oldType string, valueStr string) (newType string, changed bool) {
	if valueStr == "" {
		return oldType, false
	}

	// Определяем новый тип на основе значения
	trimmed := strings.TrimSpace(valueStr)

	// Если текущий тип - bool, то "0" и "1" остаются bool
	if oldType == "bool" {
		if trimmed == "0" || trimmed == "1" {
			return "bool", false
		}
		// Если введено что-то другое, определяем новый тип
	}

	// Если текущий тип - number, то "0" и "1" остаются number
	if oldType == "number" {
		if trimmed == "0" || trimmed == "1" {
			return "number", false
		}
		// Проверяем, является ли это числом
		if trimmed == "-" || trimmed == "." || trimmed == "-." {
			return oldType, false // Промежуточное состояние
		}
		trimmedForNumber := trimmed
		if strings.HasSuffix(trimmedForNumber, ".") {
			trimmedForNumber = strings.TrimSuffix(trimmedForNumber, ".")
		}
		_, errInt := strconv.ParseInt(trimmedForNumber, 10, 64)
		_, errFloat := strconv.ParseFloat(trimmedForNumber, 64)
		if errInt == nil || errFloat == nil {
			return "number", false // Остается number
		}
		// Если не число, то это строка
		if oldType != "string" {
			return "string", true
		}
		return "string", false
	}

	// Если текущий тип - string, определяем новый тип на основе значения
	if oldType == "string" {
		// Проверяем bool (только 0 и 1)
		if trimmed == "0" || trimmed == "1" {
			return "bool", true
		}
		// Проверяем число
		if trimmed == "-" || trimmed == "." || trimmed == "-." {
			return oldType, false // Промежуточное состояние
		}
		trimmedForNumber := trimmed
		if strings.HasSuffix(trimmedForNumber, ".") {
			trimmedForNumber = strings.TrimSuffix(trimmedForNumber, ".")
		}
		_, errInt := strconv.ParseInt(trimmedForNumber, 10, 64)
		_, errFloat := strconv.ParseFloat(trimmedForNumber, 64)
		if errInt == nil || errFloat == nil {
			return "number", true
		}
		// Остается string
		return "string", false
	}

	// Для неизвестных типов определяем тип на основе значения
	// Проверяем bool (только 0 и 1)
	if trimmed == "0" || trimmed == "1" {
		return "bool", oldType != "bool"
	}
	// Проверяем число
	if trimmed == "-" || trimmed == "." || trimmed == "-." {
		return oldType, false // Промежуточное состояние
	}
	trimmedForNumber := trimmed
	if strings.HasSuffix(trimmedForNumber, ".") {
		trimmedForNumber = strings.TrimSuffix(trimmedForNumber, ".")
	}
	_, errInt := strconv.ParseInt(trimmedForNumber, 10, 64)
	_, errFloat := strconv.ParseFloat(trimmedForNumber, 64)
	if errInt == nil || errFloat == nil {
		return "number", oldType != "number"
	}
	// Строка
	return "string", oldType != "string"
}

// showTypeChangeDialog показывает диалог подтверждения изменения типа
func (a *ProtobufTreeAdapter) showTypeChangeDialog(oldType, newType string, onConfirm func(), onCancel func()) {
	if a.window == nil {
		// Если нет окна, просто очищаем поле
		onCancel()
		return
	}

	// Если установлен флаг "больше не спрашивать", просто очищаем поле
	if dontAskTypeChangeConfirmation {
		onConfirm()
		return
	}

	// Создаем кастомный диалог с чекбоксом
	message := fmt.Sprintf("Changing field type from '%s' to '%s' will clear the field value.\n\nDo you want to continue?", oldType, newType)
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	checkbox := widget.NewCheck("Don't ask again", func(checked bool) {
		dontAskTypeChangeConfirmation = checked
	})

	content := container.NewVBox(
		label,
		checkbox,
	)

	var customDialog dialog.Dialog
	confirmBtn := widget.NewButton("Clear", func() {
		customDialog.Hide()
		onConfirm()
	})
	cancelBtn := widget.NewButton("Cancel", func() {
		customDialog.Hide()
		onCancel()
	})

	buttons := container.NewHBox(confirmBtn, cancelBtn)
	dialogContent := container.NewVBox(content, buttons)

	customDialog = dialog.NewCustom("Type Change Confirmation", "Close", dialogContent, a.window)
	customDialog.Resize(fyne.NewSize(400, 200))
	customDialog.Show()
}

// handleTypeChange обрабатывает изменение типа поля
func (a *ProtobufTreeAdapter) handleTypeChange(uid widget.TreeNodeID, oldType, newType string) {
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	// Проверяем, можно ли безшовно изменить тип
	canSeamlessChange := false
	valueStr := a.nodeValueToString(node)
	if (oldType == "bool" && newType == "number") || (oldType == "number" && newType == "bool") {
		// Проверяем, можно ли конвертировать значение (0 и 1)
		if valueStr == "0" || valueStr == "1" {
			canSeamlessChange = true
		}
	}

	// Если можно безшовно изменить, делаем это
	if canSeamlessChange {
		node.Type = newType
		// Конвертируем значение при необходимости
		if oldType == "bool" && newType == "number" {
			// bool -> number: true -> "1", false -> "0"
			// valueStr уже в формате "0" или "1" (из nodeValueToString)
			node.Value = valueStr
		} else if oldType == "number" && newType == "bool" {
			// number -> bool: "1" -> true, "0" -> false
			if valueStr == "1" {
				node.Value = true
			} else if valueStr == "0" {
				node.Value = false
			}
		}
		// Обновляем значение в виджете
		if editWidget, ok := a.editWidgets[uid]; ok {
			editWidget.typeCombo.SetSelected(newType)
			// Обновляем значение в Entry
			newValueStr := a.nodeValueToString(node)
			editWidget.entry.SetText(newValueStr)
			// Обновляем валидацию с новым типом
			a.updateEntryValidation(uid, newType)
		}
		return
	}

	// Иначе показываем диалог подтверждения
	oldValue := node.Value
	a.showTypeChangeDialog(
		oldType,
		newType,
		func() {
			// Подтверждение: очищаем поле и обновляем тип
			node.Value = nil
			node.Type = newType
			// Обновляем значение в виджете
			if editWidget, ok := a.editWidgets[uid]; ok {
				editWidget.entry.SetText("")
				editWidget.typeCombo.SetSelected(newType)
				// Обновляем валидацию с новым типом
				a.updateEntryValidation(uid, newType)
			}
		},
		func() {
			// Отмена: восстанавливаем старый тип
			node.Type = oldType
			node.Value = oldValue
			// Обновляем значение в виджете
			if editWidget, ok := a.editWidgets[uid]; ok {
				valueStr := a.nodeValueToString(node)
				editWidget.entry.SetText(valueStr)
				editWidget.typeCombo.SetSelected(oldType)
				// Восстанавливаем валидацию со старым типом
				a.updateEntryValidation(uid, oldType)
			}
		},
	)
}

// updateNodeValue обновляет значение узла
func (a *ProtobufTreeAdapter) updateNodeValue(uid widget.TreeNodeID, valueStr string, fieldType string) {
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	// Определяем, изменился ли тип
	newType, typeChanged := a.detectTypeChange(fieldType, valueStr)

	// Если тип изменился и это не безшовное изменение, показываем диалог
	if typeChanged {
		// Сохраняем текущее значение для отката
		oldValue := node.Value
		oldType := node.Type

		// Показываем диалог подтверждения
		a.showTypeChangeDialog(
			oldType,
			newType,
			func() {
				// Подтверждение: очищаем поле и обновляем тип
				node.Value = nil
				node.Type = newType
				// Обновляем значение в виджете
				if editWidget, ok := a.editWidgets[uid]; ok {
					editWidget.entry.SetText("")
					editWidget.typeCombo.SetSelected(newType)
					// Обновляем валидацию с новым типом
					a.updateEntryValidation(uid, newType)
				}
			},
			func() {
				// Отмена: восстанавливаем старое значение
				node.Value = oldValue
				node.Type = oldType
				// Обновляем значение в виджете
				if editWidget, ok := a.editWidgets[uid]; ok {
					valueStr := a.nodeValueToString(node)
					editWidget.entry.SetText(valueStr)
					editWidget.typeCombo.SetSelected(oldType)
				}
			},
		)
		return
	}

	// Тип не изменился или безшовное изменение - обновляем значение
	// Обновляем тип узла, если он изменился (для безшовных изменений)
	if newType != fieldType {
		node.Type = newType
		// Обновляем тип в виджете
		if editWidget, ok := a.editWidgets[uid]; ok {
			editWidget.typeCombo.SetSelected(newType)
		}
	}

	// Парсим значение в зависимости от типа
	switch newType {
	case "string":
		node.Value = valueStr
	case "number":
		// Для чисел сохраняем как строку (как при парсинге)
		node.Value = valueStr
	case "bool":
		// Для bool принимаем только "0" и "1"
		trimmed := strings.TrimSpace(valueStr)
		if trimmed == "1" {
			node.Value = true
		} else if trimmed == "0" {
			node.Value = false
		} else {
			// Если не распознано, сохраняем как строку (не должно происходить при валидации)
			node.Value = valueStr
		}
	default:
		node.Value = valueStr
	}
}

// getNodeByUID получает узел по UID
func (a *ProtobufTreeAdapter) getNodeByUID(uid widget.TreeNodeID) *protobuf.TreeNode {
	if a.tree == nil {
		return nil
	}

	// В Fyne widget.Tree использует пустую строку "" для root
	// Также может быть "root" или числовой формат для дочерних узлов
	if uid == "" || uid == "root" {
		return a.tree
	}

	// Парсим путь - может быть числовой формат "0", "0:1" или "root:0:1"
	var parts []string
	if strings.HasPrefix(uid, "root:") {
		// Формат "root:0:1"
		parts = splitUID(uid)
		if len(parts) == 0 || parts[0] != "root" {
			return nil
		}
	} else {
		// Числовой формат "0", "0:1" - начинаем с root
		parts = splitUID(uid)
		if len(parts) == 0 {
			return nil
		}
		// Добавляем "root" в начало
		parts = append([]string{"root"}, parts...)
	}

	// Навигация по дереву (начинаем с индекса 1, так как parts[0] = "root")
	current := a.tree
	for i := 1; i < len(parts); i++ {
		idx := parseInt(parts[i])
		if idx < 0 || idx >= len(current.Children) {
			return nil
		}
		current = current.Children[idx]
	}

	return current
}

// DebugPrintTree выводит дерево для отладки (вывод отключен)
func (a *ProtobufTreeAdapter) DebugPrintTree() {
	// Метод оставлен для совместимости, но вывод отключен
}

// splitUID разбивает UID на части
func splitUID(uid widget.TreeNodeID) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range uid {
		if char == ':' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// parseInt парсит строку в число
func parseInt(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

// CreateProtobufTree создает виджет дерева для protobuf
func CreateProtobufTree(tree *protobuf.TreeNode) *widget.Tree {
	if tree == nil {
		// Возвращаем пустое дерево
		adapter := NewProtobufTreeAdapter(&protobuf.TreeNode{
			Name:     "root",
			Type:     "message",
			Children: make([]*protobuf.TreeNode, 0),
		})
		treeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
		return treeWidget
	}

	adapter := NewProtobufTreeAdapter(tree)
	treeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)

	// Разворачиваем root по умолчанию
	treeWidget.OpenBranch("root")

	return treeWidget
}
