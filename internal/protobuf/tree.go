package protobuf

import "strings"

type TreeNode struct {
	Name       string
	Type       string
	Value      interface{}
	Children   []*TreeNode
	FieldNum   int
	IsRepeated bool
}

func NewTreeNode(name, fieldType string, fieldNum int) *TreeNode {
	return &TreeNode{
		Name:     name,
		Type:     fieldType,
		FieldNum: fieldNum,
		Children: make([]*TreeNode, 0),
	}
}

func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

func (n *TreeNode) IsMessage() bool {
	return len(n.Children) > 0 || isMessageType(n.Type)
}

func isMessageType(t string) bool {
	return t == "message" || strings.HasPrefix(t, "message_")
}
