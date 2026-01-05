package protobuf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseRawWithNestedMessage(t *testing.T) {
	// Тест с реальным файлом person.bin, который содержит вложенные сообщения
	testdataDir := "testdata"
	personBin := filepath.Join(testdataDir, "person.bin")

	// Получаем абсолютный путь
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Если мы в internal/protobuf, поднимаемся на два уровня вверх
	if filepath.Base(wd) == "protobuf" {
		testdataDir = filepath.Join("..", "..", "testdata")
	}
	personBin = filepath.Join(testdataDir, "person.bin")

	// Получаем абсолютный путь
	absPath, err := filepath.Abs(personBin)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Проверяем, существует ли файл
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Skipping test: %s does not exist (tried %s from %s)", absPath, personBin, wd)
		return
	}

	// Проверяем protoc
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Читаем файл
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", absPath, err)
	}

	// Используем ParseRaw (вызывает protoc --decode_raw)
	tree, err := parser.ParseRaw(data)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", absPath, err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	// Проверяем структуру дерева
	if tree.Name != "root" {
		t.Errorf("Expected root name to be 'root', got '%s'", tree.Name)
	}

	// Выводим структуру дерева для отладки
	t.Logf("\nTree structure from ParseRaw:")
	printTree(t, tree, 0)

	// Проверяем количество детей root - должно быть 6 (field_1, field_2, field_3, field_4, field_5, field_5)
	expectedRootChildren := 6
	if len(tree.Children) != expectedRootChildren {
		t.Errorf("Expected root to have %d children, got %d", expectedRootChildren, len(tree.Children))
		t.Logf("Root children:")
		for i, child := range tree.Children {
			t.Logf("  [%d] %s (field_%d, %s) - children: %d", i, child.Name, child.FieldNum, child.Type, len(child.Children))
		}
	}

	// Находим field_4 (Address message) - это должен быть 4-й элемент (индекс 3)
	var field4 *TreeNode
	field4Index := -1
	for i, child := range tree.Children {
		if child.FieldNum == 4 {
			field4 = child
			field4Index = i
			break
		}
	}

	if field4 == nil {
		t.Fatal("field_4 (Address message) not found in root children")
		t.Logf("Root children:")
		for i, child := range tree.Children {
			t.Logf("  [%d] %s (field_%d, %s) - children: %d", i, child.Name, child.FieldNum, child.Type, len(child.Children))
		}
	}

	if field4.Type != "message" {
		t.Errorf("Expected field_4 to be of type 'message', got '%s'", field4.Type)
	}

	// ВАЖНО: Проверяем, что field_4 имеет 4 детей (street, city, country, zip_code)
	expectedField4Children := 4
	if len(field4.Children) != expectedField4Children {
		t.Errorf("Expected field_4 to have %d children, got %d", expectedField4Children, len(field4.Children))
		t.Logf("field_4 children:")
		for i, child := range field4.Children {
			t.Logf("  [%d] %s (field_%d, %s) = %v", i, child.Name, child.FieldNum, child.Type, child.Value)
		}
		t.Logf("\nFull tree structure:")
		printTree(t, tree, 0)
	}

	// Проверяем, что все дети field_4 находятся на правильных позициях
	// field_4 должен быть на индексе 3 в root.Children
	if field4Index != 3 {
		t.Errorf("Expected field_4 to be at index 3 in root.Children, got index %d", field4Index)
	}

	// Проверяем, что дети field_4 находятся ВНУТРИ field_4, а не в root
	for i, child := range tree.Children {
		if i != field4Index && child.FieldNum >= 1 && child.FieldNum <= 4 {
			// Это может быть проблемой - если мы находим field_1-4 вне field_4, это неправильно
			// Но нужно быть осторожным, так как field_1-3 могут быть и в root
			if i > field4Index {
				t.Errorf("Found field_%d at index %d (after field_4 at index %d) - this might indicate parsing issue", child.FieldNum, i, field4Index)
			}
		}
	}

	// Проверяем содержимое field_4 (ожидаемые значения из person.bin)
	expectedFields := []struct {
		fieldNum int
		typeName string
		value    interface{}
	}{
		{1, "string", "123 Main St"},
		{2, "string", "New York"},
		{3, "string", "USA"},
		{4, "number", "10001"},
	}

	if len(field4.Children) >= len(expectedFields) {
		for i, expected := range expectedFields {
			if i >= len(field4.Children) {
				t.Errorf("field_4.Children[%d]: expected field, got nothing", i)
				continue
			}
			child := field4.Children[i]
			if child.FieldNum != expected.fieldNum {
				t.Errorf("field_4.Children[%d]: expected FieldNum=%d, got %d", i, expected.fieldNum, child.FieldNum)
			}
			if child.Type != expected.typeName {
				t.Errorf("field_4.Children[%d]: expected Type='%s', got '%s'", i, expected.typeName, child.Type)
			}
			if child.Value != expected.value {
				t.Errorf("field_4.Children[%d]: expected Value=%v, got %v", i, expected.value, child.Value)
			}
		}
	}
}

// printTree выводит дерево для отладки
func printTree(t *testing.T, node *TreeNode, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	valueStr := ""
	if node.Value != nil {
		if s, ok := node.Value.(string); ok {
			valueStr = " = \"" + s + "\""
		} else {
			valueStr = " = " + stringifyValue(node.Value)
		}
	}
	t.Logf("%s%s (field_%d, %s)%s [children: %d]", prefix, node.Name, node.FieldNum, node.Type, valueStr, len(node.Children))
	for _, child := range node.Children {
		printTree(t, child, indent+1)
	}
}

func stringifyValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return "\"" + s + "\""
	}
	return ""
}
