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

// CreateFakeTree создает фейковое дерево с правильной иерархией для тестирования виджета
func CreateFakeTree() *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	// Добавляем простые поля в root
	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "John Doe",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "30",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	field3 := &TreeNode{
		Name:     "field_3",
		Type:     "string",
		FieldNum: 3,
		Value:    "john@example.com",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field3)

	// Создаем вложенное сообщение field_4 (Address)
	field4 := &TreeNode{
		Name:     "field_4",
		Type:     "message",
		FieldNum: 4,
		Children: make([]*TreeNode, 0),
	}

	// Добавляем детей в field_4
	addressField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "123 Main St",
		Children: make([]*TreeNode, 0),
	}
	field4.AddChild(addressField1)

	addressField2 := &TreeNode{
		Name:     "field_2",
		Type:     "string",
		FieldNum: 2,
		Value:    "New York",
		Children: make([]*TreeNode, 0),
	}
	field4.AddChild(addressField2)

	addressField3 := &TreeNode{
		Name:     "field_3",
		Type:     "string",
		FieldNum: 3,
		Value:    "USA",
		Children: make([]*TreeNode, 0),
	}
	field4.AddChild(addressField3)

	addressField4 := &TreeNode{
		Name:     "field_4",
		Type:     "number",
		FieldNum: 4,
		Value:    "10001",
		Children: make([]*TreeNode, 0),
	}
	field4.AddChild(addressField4)

	// Добавляем field_4 в root
	root.AddChild(field4)

	// Добавляем еще одно простое поле
	field5 := &TreeNode{
		Name:     "field_5",
		Type:     "string",
		FieldNum: 5,
		Value:    "reading",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field5)

	field5_2 := &TreeNode{
		Name:     "field_5",
		Type:     "string",
		FieldNum: 5,
		Value:    "coding",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field5_2)

	return root
}
