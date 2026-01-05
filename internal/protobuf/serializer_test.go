package protobuf

import (
	"os/exec"
	"strings"
	"testing"
)

// TestGenerateProtoSchema_SimpleFields тестирует генерацию proto схемы для простых полей
func TestGenerateProtoSchema_SimpleFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем простое дерево с примитивными полями
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	field3 := &TreeNode{
		Name:     "field_3",
		Type:     "bool",
		FieldNum: 3,
		Value:    true,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field3)

	// Генерируем схему
	schema := parser.generateProtoSchema(root)

	// Проверяем базовую структуру
	if !strings.Contains(schema, "syntax = \"proto3\";") {
		t.Error("Schema should contain syntax declaration")
	}

	if !strings.Contains(schema, "message Message {") {
		t.Error("Schema should contain Message declaration")
	}

	// Проверяем наличие полей
	if !strings.Contains(schema, "string field_1 = 1;") {
		t.Errorf("Schema should contain string field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "int32 field_2 = 2;") {
		t.Errorf("Schema should contain int32 field_2 = 2, got:\n%s", schema)
	}

	if !strings.Contains(schema, "bool field_3 = 3;") {
		t.Errorf("Schema should contain bool field_3 = 3, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

// TestGenerateProtoSchema_NestedMessage тестирует генерацию proto схемы с вложенными сообщениями
func TestGenerateProtoSchema_NestedMessage(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем дерево с вложенным сообщением
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	// Вложенное сообщение
	nestedMessage := &TreeNode{
		Name:     "field_2",
		Type:     "message",
		FieldNum: 2,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "nested_value",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField1)

	nestedField2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "100",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField2)

	root.AddChild(nestedMessage)

	// Генерируем схему
	schema := parser.generateProtoSchema(root)

	// Проверяем наличие вложенного сообщения
	if !strings.Contains(schema, "Message1 field_2 = 2;") {
		t.Errorf("Schema should contain Message1 field_2 = 2, got:\n%s", schema)
	}

	if !strings.Contains(schema, "message Message1 {") {
		t.Errorf("Schema should contain nested message Message1, got:\n%s", schema)
	}

	// Проверяем поля внутри вложенного сообщения
	if !strings.Contains(schema, "string field_1 = 1;") {
		t.Errorf("Schema should contain nested string field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "int32 field_2 = 2;") {
		t.Errorf("Schema should contain nested int32 field_2 = 2, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

// TestGenerateProtoSchema_RepeatedFields тестирует генерацию proto схемы для repeated полей
func TestGenerateProtoSchema_RepeatedFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем дерево с repeated полем
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	// Добавляем одно и то же поле несколько раз (repeated)
	field1 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value1",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field1_2 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value2",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1_2)

	// Генерируем схему
	schema := parser.generateProtoSchema(root)

	// Проверяем, что поле помечено как repeated
	if !strings.Contains(schema, "repeated string field_1 = 1;") {
		t.Errorf("Schema should contain 'repeated string field_1 = 1', got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

// TestMapTypeToProtoType тестирует преобразование типов
func TestMapTypeToProtoType(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"number", "int32"},
		{"bool", "bool"},
		{"unknown", "string"}, // По умолчанию string
	}

	for _, tt := range tests {
		result := parser.mapTypeToProtoType(tt.input)
		if result != tt.expected {
			t.Errorf("mapTypeToProtoType(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestTreeToTextFormatWithNames_SimpleFields тестирует генерацию текстового формата для простых полей
func TestTreeToTextFormatWithNames_SimpleFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем простое дерево
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "Hello, World!",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	field3 := &TreeNode{
		Name:     "field_3",
		Type:     "bool",
		FieldNum: 3,
		Value:    true,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field3)

	// Генерируем текстовый формат
	textFormat := parser.treeToTextFormatWithNames(root)

	// Проверяем наличие полей с именами
	if !strings.Contains(textFormat, "field_1: \"Hello, World!\"") {
		t.Errorf("Text format should contain field_1 with value, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "field_2: 42") {
		t.Errorf("Text format should contain field_2 with value, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "field_3: true") {
		t.Errorf("Text format should contain field_3 with value, got:\n%s", textFormat)
	}

	t.Logf("Generated text format:\n%s", textFormat)
}

// TestTreeToTextFormatWithNames_NestedMessage тестирует генерацию текстового формата с вложенными сообщениями
func TestTreeToTextFormatWithNames_NestedMessage(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем дерево с вложенным сообщением
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	// Вложенное сообщение
	nestedMessage := &TreeNode{
		Name:     "field_2",
		Type:     "message",
		FieldNum: 2,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "nested_value",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField1)

	root.AddChild(nestedMessage)

	// Генерируем текстовый формат
	textFormat := parser.treeToTextFormatWithNames(root)

	// Проверяем наличие вложенного сообщения
	if !strings.Contains(textFormat, "field_2 {") {
		t.Errorf("Text format should contain field_2 message, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "field_1: \"nested_value\"") {
		t.Errorf("Text format should contain nested field_1, got:\n%s", textFormat)
	}

	t.Logf("Generated text format:\n%s", textFormat)
}

// TestTreeToTextFormatWithNames_RepeatedFields тестирует генерацию текстового формата для repeated полей
func TestTreeToTextFormatWithNames_RepeatedFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Создаем дерево с repeated полем
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	// Добавляем одно и то же поле несколько раз
	field1 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value1",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field1_2 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value2",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1_2)

	field1_3 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value3",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1_3)

	// Генерируем текстовый формат
	textFormat := parser.treeToTextFormatWithNames(root)

	// Проверяем, что все значения repeated поля присутствуют
	if !strings.Contains(textFormat, "field_1: \"value1\"") {
		t.Errorf("Text format should contain field_1: \"value1\", got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "field_1: \"value2\"") {
		t.Errorf("Text format should contain field_1: \"value2\", got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "field_1: \"value3\"") {
		t.Errorf("Text format should contain field_1: \"value3\", got:\n%s", textFormat)
	}

	t.Logf("Generated text format:\n%s", textFormat)
}

// TestSerializeRaw_RoundTrip тестирует полный цикл: парсинг -> сериализация -> парсинг
func TestSerializeRaw_RoundTrip(t *testing.T) {
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

	// Создаем тестовое дерево
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "Hello, World!",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	field3 := &TreeNode{
		Name:     "field_3",
		Type:     "bool",
		FieldNum: 3,
		Value:    true,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field3)

	// Сериализуем дерево
	binaryData, err := parser.SerializeRaw(root)
	if err != nil {
		t.Fatalf("Failed to serialize tree: %v", err)
	}

	if len(binaryData) == 0 {
		t.Fatal("Serialized data is empty")
	}

	// Парсим обратно
	parsedTree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse serialized data: %v", err)
	}

	if parsedTree == nil {
		t.Fatal("Parsed tree is nil")
	}

	// Проверяем структуру
	if len(parsedTree.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(parsedTree.Children))
	}

	// Проверяем значения (номера полей должны совпадать)
	expectedFields := map[int]struct {
		fieldNum int
		typeName string
	}{
		0: {1, "string"},
		1: {2, "number"},
		2: {3, "bool"},
	}

	for i, expected := range expectedFields {
		if i >= len(parsedTree.Children) {
			t.Errorf("Child %d not found", i)
			continue
		}

		child := parsedTree.Children[i]
		if child.FieldNum != expected.fieldNum {
			t.Errorf("Child %d: expected FieldNum=%d, got %d", i, expected.fieldNum, child.FieldNum)
		}

		if child.Type != expected.typeName {
			t.Errorf("Child %d: expected Type=%s, got %s", i, expected.typeName, child.Type)
		}
	}

	t.Logf("Round trip successful: serialized %d bytes, parsed back to %d children", len(binaryData), len(parsedTree.Children))
}

// TestSerializeRaw_WithNestedMessage тестирует сериализацию с вложенными сообщениями
func TestSerializeRaw_WithNestedMessage(t *testing.T) {
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

	// Создаем дерево с вложенным сообщением
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	// Вложенное сообщение
	nestedMessage := &TreeNode{
		Name:     "field_2",
		Type:     "message",
		FieldNum: 2,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "nested_value",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField1)

	nestedField2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "100",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField2)

	root.AddChild(nestedMessage)

	// Сериализуем дерево
	binaryData, err := parser.SerializeRaw(root)
	if err != nil {
		t.Fatalf("Failed to serialize tree: %v", err)
	}

	if len(binaryData) == 0 {
		t.Fatal("Serialized data is empty")
	}

	// Парсим обратно
	parsedTree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse serialized data: %v", err)
	}

	if parsedTree == nil {
		t.Fatal("Parsed tree is nil")
	}

	// Проверяем структуру
	if len(parsedTree.Children) < 2 {
		t.Errorf("Expected at least 2 children, got %d", len(parsedTree.Children))
	}

	// Ищем вложенное сообщение
	var nestedMsg *TreeNode
	for _, child := range parsedTree.Children {
		if child.FieldNum == 2 && child.Type == "message" {
			nestedMsg = child
			break
		}
	}

	if nestedMsg == nil {
		t.Fatal("Nested message not found in parsed tree")
	}

	if len(nestedMsg.Children) < 2 {
		t.Errorf("Expected nested message to have at least 2 children, got %d", len(nestedMsg.Children))
	}

	t.Logf("Round trip with nested message successful: serialized %d bytes", len(binaryData))
}

// TestSerializeRaw_WithRepeatedFields тестирует сериализацию с repeated полями
func TestSerializeRaw_WithRepeatedFields(t *testing.T) {
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

	// Создаем дерево с repeated полем
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	// Добавляем одно и то же поле несколько раз
	field1 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value1",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field1_2 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "value2",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1_2)

	// Сериализуем дерево
	binaryData, err := parser.SerializeRaw(root)
	if err != nil {
		t.Fatalf("Failed to serialize tree: %v", err)
	}

	if len(binaryData) == 0 {
		t.Fatal("Serialized data is empty")
	}

	// Парсим обратно
	parsedTree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse serialized data: %v", err)
	}

	if parsedTree == nil {
		t.Fatal("Parsed tree is nil")
	}

	// Проверяем, что repeated поля были сохранены
	// В protobuf repeated поля могут быть представлены как несколько узлов с одинаковым FieldNum
	field1Count := 0
	for _, child := range parsedTree.Children {
		if child.FieldNum == 1 {
			field1Count++
		}
	}

	if field1Count < 2 {
		t.Errorf("Expected at least 2 instances of field_1, got %d", field1Count)
	}

	t.Logf("Round trip with repeated fields successful: serialized %d bytes, found %d instances of field_1", len(binaryData), field1Count)
}

