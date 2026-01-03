package protobuf

// TreeNode представляет узел в дереве protobuf
type TreeNode struct {
	Name       string      // Имя поля
	Type       string      // Тип поля (string, int32, message, etc.)
	Value      interface{} // Значение (для примитивных типов) или nil для сообщений
	Children   []*TreeNode // Дочерние узлы (для сообщений)
	FieldNum   int         // Номер поля в protobuf
	IsRepeated bool        // Является ли поле повторяющимся (repeated)
}

// NewTreeNode создает новый узел дерева
func NewTreeNode(name, fieldType string, fieldNum int) *TreeNode {
	return &TreeNode{
		Name:     name,
		Type:     fieldType,
		FieldNum: fieldNum,
		Children: make([]*TreeNode, 0),
	}
}

// AddChild добавляет дочерний узел
func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

// IsMessage возвращает true, если узел является сообщением (имеет дочерние элементы)
func (n *TreeNode) IsMessage() bool {
	return len(n.Children) > 0 || n.Type == "message"
}
