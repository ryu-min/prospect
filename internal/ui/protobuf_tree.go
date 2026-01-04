package ui

import (
	"fmt"
	"os"
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
		fmt.Fprintf(os.Stdout, "[DEBUG] ChildUIDs: узел не найден для UID: '%s' (actualUID: '%s')\n", uid, actualUID)
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
	fmt.Fprintf(os.Stdout, "[DEBUG] ChildUIDs: UID='%s' (actualUID='%s'), children=%v\n", uid, actualUID, children)
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
	isBranch := len(node.Children) > 0
	fmt.Fprintf(os.Stdout, "[DEBUG] IsBranch: UID='%s' (actualUID='%s'), isBranch=%v\n", uid, actualUID, isBranch)
	return isBranch
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
		fmt.Fprintf(os.Stdout, "[DEBUG] UpdateNode: узел не найден для UID: '%s' (actualUID: '%s')\n", uid, actualUID)
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
	fmt.Fprintf(os.Stdout, "[DEBUG] UpdateNode: UID='%s' (actualUID='%s'), text=%s\n", uid, actualUID, text)
}

// getNodeByUID получает узел по UID
func (a *ProtobufTreeAdapter) getNodeByUID(uid widget.TreeNodeID) *protobuf.TreeNode {
	if a.tree == nil {
		fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: tree is nil\n")
		return nil
	}

	// В Fyne widget.Tree использует пустую строку "" для root
	// Также может быть "root" или числовой формат для дочерних узлов
	if uid == "" || uid == "root" {
		fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: возвращаем root, children=%d\n", len(a.tree.Children))
		return a.tree
	}

	// Парсим путь - может быть числовой формат "0", "0:1" или "root:0:1"
	var parts []string
	if strings.HasPrefix(uid, "root:") {
		// Формат "root:0:1"
		parts = splitUID(uid)
		if len(parts) == 0 || parts[0] != "root" {
			fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: неверный формат UID (root:...)\n")
			return nil
		}
	} else {
		// Числовой формат "0", "0:1" - начинаем с root
		parts = splitUID(uid)
		if len(parts) == 0 {
			fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: пустой UID после парсинга\n")
			return nil
		}
		// Добавляем "root" в начало
		parts = append([]string{"root"}, parts...)
	}

	fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: UID='%s', parts=%v\n", uid, parts)

	// Навигация по дереву (начинаем с индекса 1, так как parts[0] = "root")
	current := a.tree
	for i := 1; i < len(parts); i++ {
		idx := parseInt(parts[i])
		fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: часть %d, idx=%d, len(children)=%d\n", i, idx, len(current.Children))
		if idx < 0 || idx >= len(current.Children) {
			fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: индекс вне диапазона\n")
			return nil
		}
		current = current.Children[idx]
	}

	fmt.Fprintf(os.Stdout, "[DEBUG] getNodeByUID: найден узел %s\n", current.Name)
	return current
}

// DebugPrintTree выводит дерево для отладки
func (a *ProtobufTreeAdapter) DebugPrintTree() {
	fmt.Fprintf(os.Stdout, "[DEBUG] Tree structure:\n")
	a.printNode(a.tree, 0)
}

func (a *ProtobufTreeAdapter) printNode(node *protobuf.TreeNode, indent int) {
	if node == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)
	fmt.Fprintf(os.Stdout, "%s%s (field_%d, %s)", prefix, node.Name, node.FieldNum, node.Type)
	if node.Value != nil {
		fmt.Fprintf(os.Stdout, " = %v", node.Value)
	}
	fmt.Fprintf(os.Stdout, " [children: %d]\n", len(node.Children))
	for _, child := range node.Children {
		a.printNode(child, indent+1)
	}
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
