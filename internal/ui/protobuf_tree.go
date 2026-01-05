package ui

import (
	"fmt"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// ProtobufTreeAdapter адаптирует TreeNode для widget.Tree
type ProtobufTreeAdapter struct {
	tree        *protobuf.TreeNode
	editWidgets map[widget.TreeNodeID]*EditableNodeWidget
}

// EditableNodeWidget - виджет для редактирования значения узла
type EditableNodeWidget struct {
	widget.BaseWidget
	nameLabel *widget.Label // Имя поля
	typeLabel *widget.Label // Тип поля
	entry     *widget.Entry // Значение для редактирования
	uid       widget.TreeNodeID
	adapter   *ProtobufTreeAdapter
}

// Константы для ширины колонок
const (
	nameColumnWidth = 120 // Ширина колонки имени
	typeColumnWidth = 80  // Ширина колонки типа
	columnSpacing   = 10  // Отступ между колонками
)

// NewEditableNodeWidget создает новый редактируемый виджет узла
func NewEditableNodeWidget(uid widget.TreeNodeID, adapter *ProtobufTreeAdapter) *EditableNodeWidget {
	ew := &EditableNodeWidget{
		uid:       uid,
		adapter:   adapter,
		nameLabel: widget.NewLabel(""),
		typeLabel: widget.NewLabel(""),
		entry:     widget.NewEntry(),
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
		typeLabel: ew.typeLabel,
		entry:     ew.entry,
		objects:   []fyne.CanvasObject{ew.nameLabel, ew.typeLabel, ew.entry},
	}
}

type editableNodeRenderer struct {
	widget    *EditableNodeWidget
	nameLabel *widget.Label
	typeLabel *widget.Label
	entry     *widget.Entry
	objects   []fyne.CanvasObject
}

func (r *editableNodeRenderer) Layout(size fyne.Size) {
	// Все колонки начинаются на одном уровне (с учетом отступов дерева)
	// Колонка имени
	namePos := fyne.NewPos(0, (size.Height-r.nameLabel.MinSize().Height)/2)
	r.nameLabel.Move(namePos)
	r.nameLabel.Resize(fyne.NewSize(float32(nameColumnWidth), r.nameLabel.MinSize().Height))

	// Колонка типа
	typePos := fyne.NewPos(float32(nameColumnWidth+columnSpacing), (size.Height-r.typeLabel.MinSize().Height)/2)
	r.typeLabel.Move(typePos)
	r.typeLabel.Resize(fyne.NewSize(float32(typeColumnWidth), r.typeLabel.MinSize().Height))

	// Колонка значения (Entry) - занимает оставшееся пространство
	entryX := float32(nameColumnWidth + typeColumnWidth + columnSpacing*2)
	entryWidth := size.Width - entryX
	entryPos := fyne.NewPos(entryX, (size.Height-r.entry.MinSize().Height)/2)
	r.entry.Move(entryPos)
	r.entry.Resize(fyne.NewSize(entryWidth, r.entry.MinSize().Height))
}

func (r *editableNodeRenderer) MinSize() fyne.Size {
	nameSize := r.nameLabel.MinSize()
	typeSize := r.typeLabel.MinSize()
	entrySize := r.entry.MinSize()

	width := float32(nameColumnWidth + typeColumnWidth + int(entrySize.Width) + columnSpacing*2)
	height := fyne.Max(fyne.Max(nameSize.Height, typeSize.Height), entrySize.Height)

	return fyne.NewSize(width, height)
}

func (r *editableNodeRenderer) Refresh() {
	r.nameLabel.Refresh()
	r.typeLabel.Refresh()
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
	}
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

		// Устанавливаем тип поля во вторую колонку
		typeText := node.Type
		editWidget.typeLabel.SetText(typeText)

		// Получаем текущее значение из узла (источник истины)
		// ВАЖНО: Всегда читаем из узла, а не из виджета
		valueStr := a.nodeValueToString(node)

		// Временно отключаем OnChanged, чтобы не триггерить обновление при SetText
		editWidget.entry.OnChanged = nil
		editWidget.entry.SetText(valueStr)

		// Устанавливаем обработчик с правильным типом после установки значения
		fieldType := node.Type
		editWidget.entry.OnChanged = func(value string) {
			// Всегда обновляем значение в узле при изменении
			// Узел - это источник истины, виджет только отображает его
			a.updateNodeValue(actualUID, value, fieldType)
		}
		editWidget.Refresh()
		return
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
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// updateNodeValue обновляет значение узла
func (a *ProtobufTreeAdapter) updateNodeValue(uid widget.TreeNodeID, valueStr string, fieldType string) {
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	// Парсим значение в зависимости от типа
	switch fieldType {
	case "string":
		node.Value = valueStr
	case "number":
		// Для чисел сохраняем как строку (как при парсинге)
		node.Value = valueStr
	case "bool":
		if valueStr == "true" || valueStr == "1" {
			node.Value = true
		} else if valueStr == "false" || valueStr == "0" {
			node.Value = false
		} else {
			// Если не распознано, сохраняем как строку
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
