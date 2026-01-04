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
	node := a.getNodeByUID(uid)
	if node == nil {
		return nil
	}

	children := make([]widget.TreeNodeID, 0, len(node.Children))
	for i := range node.Children {
		childUID := fmt.Sprintf("%s:%d", uid, i)
		children = append(children, childUID)
	}
	return children
}

// IsBranch возвращает true, если узел является веткой (имеет детей)
func (a *ProtobufTreeAdapter) IsBranch(uid widget.TreeNodeID) bool {
	node := a.getNodeByUID(uid)
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
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	label := obj.(*widget.Label)

	// Формируем текст для отображения
	text := ""
	if node.Name != "root" {
		text = fmt.Sprintf("%s (field_%d, %s)", node.Name, node.FieldNum, node.Type)
		if node.Value != nil {
			text += fmt.Sprintf(": %v", node.Value)
		}
		if node.IsRepeated {
			text += " [repeated]"
		}
	} else {
		text = "root"
		if len(node.Children) == 0 {
			text = "root (нет данных)"
		}
	}

	label.SetText(text)
}

// getNodeByUID получает узел по UID
func (a *ProtobufTreeAdapter) getNodeByUID(uid widget.TreeNodeID) *protobuf.TreeNode {
	if a.tree == nil {
		return nil
	}

	// UID имеет формат "root" или "root:0:1:2" для вложенных узлов
	if uid == "root" {
		return a.tree
	}

	// Парсим путь
	parts := splitUID(uid)
	if len(parts) == 0 || parts[0] != "root" {
		return nil
	}

	// Навигация по дереву
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
