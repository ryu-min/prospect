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
	tree *protobuf.TreeNode
}

// NewProtobufTreeAdapter создает новый адаптер для дерева
func NewProtobufTreeAdapter(tree *protobuf.TreeNode) *ProtobufTreeAdapter {
	return &ProtobufTreeAdapter{tree: tree}
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
	// Создаем контейнер с информацией о поле
	label := widget.NewLabel("")
	label.Wrapping = fyne.TextWrapWord
	return label
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

	label := obj.(*widget.Label)

	// Формируем текст для отображения
	text := ""
	if node.Name == "root" {
		text = "Protobuf Root"
		if len(node.Children) == 0 {
			text = "Protobuf Root (нет данных)"
		}
	} else {
		text = fmt.Sprintf("%s (field_%d, %s)", node.Name, node.FieldNum, node.Type)
		if node.Value != nil {
			text += fmt.Sprintf(": %v", node.Value)
		}
		if node.IsRepeated {
			text += " [repeated]"
		}
		if len(node.Children) > 0 {
			text += fmt.Sprintf(" [%d children]", len(node.Children))
		}
	}

	label.SetText(text)
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
