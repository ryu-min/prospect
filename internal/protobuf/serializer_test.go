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

	serializer := NewSerializer(parser.GetProtocPath())

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
	schema := serializer.GenerateProtoSchema(root)

	// Проверяем базовую структуру
	if !strings.Contains(schema, "syntax = \"proto2\";") {
		t.Error("Schema should contain syntax declaration")
	}

	if !strings.Contains(schema, "message Message {") {
		t.Error("Schema should contain Message declaration")
	}

	// Проверяем наличие полей
	if !strings.Contains(schema, "optional string field_1 = 1;") {
		t.Errorf("Schema should contain optional string field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "optional int64 field_2 = 2;") {
		t.Errorf("Schema should contain optional int64 field_2 = 2, got:\n%s", schema)
	}

	if !strings.Contains(schema, "optional bool field_3 = 3;") {
		t.Errorf("Schema should contain optional bool field_3 = 3, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

// TestGenerateProtoSchema_NestedMessage тестирует генерацию proto схемы с вложенными сообщениями
func TestGenerateProtoSchema_NestedMessage(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

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
	schema := serializer.GenerateProtoSchema(root)

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

	if !strings.Contains(schema, "int64 field_2 = 2;") {
		t.Errorf("Schema should contain nested int64 field_2 = 2, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

// TestGenerateProtoSchema_RepeatedFields тестирует генерацию proto схемы для repeated полей
func TestGenerateProtoSchema_RepeatedFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

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
	schema := serializer.GenerateProtoSchema(root)

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

	serializer := NewSerializer(parser.GetProtocPath())

	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"number", "int64"},
		{"bool", "bool"},
		{"unknown", "string"}, // По умолчанию string
	}

	for _, tt := range tests {
		result := serializer.MapTypeToProtoType(tt.input)
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

	serializer := NewSerializer(parser.GetProtocPath())

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
	textFormat := serializer.TreeToTextFormatWithNames(root)

	if !strings.Contains(textFormat, "1: \"Hello, World!\"") {
		t.Errorf("Text format should contain field 1 with value, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "2: 42") {
		t.Errorf("Text format should contain field 2 with value, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "3: true") {
		t.Errorf("Text format should contain field 3 with value, got:\n%s", textFormat)
	}

	t.Logf("Generated text format:\n%s", textFormat)
}

// TestTreeToTextFormatWithNames_NestedMessage тестирует генерацию текстового формата с вложенными сообщениями
func TestTreeToTextFormatWithNames_NestedMessage(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

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
	textFormat := serializer.TreeToTextFormatWithNames(root)

	if !strings.Contains(textFormat, "2 {") {
		t.Errorf("Text format should contain field 2 message, got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "1: \"nested_value\"") {
		t.Errorf("Text format should contain nested field 1, got:\n%s", textFormat)
	}

	t.Logf("Generated text format:\n%s", textFormat)
}

// TestTreeToTextFormatWithNames_RepeatedFields тестирует генерацию текстового формата для repeated полей
func TestTreeToTextFormatWithNames_RepeatedFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

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
	textFormat := serializer.TreeToTextFormatWithNames(root)

	if !strings.Contains(textFormat, "1: \"value1\"") {
		t.Errorf("Text format should contain 1: \"value1\", got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "1: \"value2\"") {
		t.Errorf("Text format should contain 1: \"value2\", got:\n%s", textFormat)
	}

	if !strings.Contains(textFormat, "1: \"value3\"") {
		t.Errorf("Text format should contain 1: \"value3\", got:\n%s", textFormat)
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

	serializer := NewSerializer(parser.GetProtocPath())

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
	binaryData, err := serializer.SerializeRaw(root)
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

	serializer := NewSerializer(parser.GetProtocPath())

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
	binaryData, err := serializer.SerializeRaw(root)
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
		if child.FieldNum == 2 && strings.HasPrefix(child.Type, "message_") {
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

	serializer := NewSerializer(parser.GetProtocPath())

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
	binaryData, err := serializer.SerializeRaw(root)
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

func TestGenerateProtoSchema_DuplicateMessageNames(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	message1 := &TreeNode{
		Name:     "message_1",
		Type:     "message",
		FieldNum: 1,
		Children: make([]*TreeNode, 0),
	}

	child1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "value1",
		Children: make([]*TreeNode, 0),
	}
	message1.AddChild(child1)

	child2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	message1.AddChild(child2)

	root.AddChild(message1)

	message2 := &TreeNode{
		Name:     "message_1",
		Type:     "message",
		FieldNum: 2,
		Children: make([]*TreeNode, 0),
	}

	child3 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "value1",
		Children: make([]*TreeNode, 0),
	}
	message2.AddChild(child3)

	child4 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	message2.AddChild(child4)

	root.AddChild(message2)

	schema := serializer.GenerateProtoSchema(root)

	if strings.Count(schema, "message Message1 {") != 1 {
		t.Errorf("Expected exactly 1 definition of Message1, got %d occurrences", strings.Count(schema, "message Message1 {"))
	}

	if !strings.Contains(schema, "Message1 message_1 = 1;") {
		t.Errorf("Schema should contain Message1 message_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "Message1 message_1_2 = 2;") {
		t.Errorf("Schema should contain Message1 message_1_2 = 2, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestRenumberMessages(t *testing.T) {
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	message4 := &TreeNode{
		Name:     "field_4",
		Type:     "message",
		FieldNum: 4,
		Children: make([]*TreeNode, 0),
	}

	message7 := &TreeNode{
		Name:     "field_7",
		Type:     "message",
		FieldNum: 7,
		Children: make([]*TreeNode, 0),
	}

	root.AddChild(message4)
	root.AddChild(message7)

	renumberMessages(root)

	if message4.Name != "field_4" {
		t.Errorf("Expected field_4 name to remain field_4, got %s", message4.Name)
	}

	if message4.Type != "message_1" {
		t.Errorf("Expected field_4 type to be message_1, got %s", message4.Type)
	}

	if message7.Name != "field_7" {
		t.Errorf("Expected field_7 name to remain field_7, got %s", message7.Name)
	}

	if message7.Type != "message_2" {
		t.Errorf("Expected field_7 type to be message_2, got %s", message7.Type)
	}

	if root.Name != "root" {
		t.Errorf("Root name should not change, got %s", root.Name)
	}
}

func TestRenumberMessages_Nested(t *testing.T) {
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	message3 := &TreeNode{
		Name:     "field_3",
		Type:     "message",
		FieldNum: 3,
		Children: make([]*TreeNode, 0),
	}

	nestedMessage5 := &TreeNode{
		Name:     "field_5",
		Type:     "message",
		FieldNum: 5,
		Children: make([]*TreeNode, 0),
	}

	message3.AddChild(nestedMessage5)
	root.AddChild(message3)

	renumberMessages(root)

	if message3.Name != "field_3" {
		t.Errorf("Expected field_3 name to remain field_3, got %s", message3.Name)
	}

	if message3.Type != "message_1" {
		t.Errorf("Expected field_3 type to be message_1, got %s", message3.Type)
	}

	if nestedMessage5.Name != "field_5" {
		t.Errorf("Expected nested field_5 name to remain field_5, got %s", nestedMessage5.Name)
	}

	if nestedMessage5.Type != "message_2" {
		t.Errorf("Expected nested field_5 type to be message_2, got %s", nestedMessage5.Type)
	}
}

func TestGenerateProtoSchema_DuplicateMessageNames_ValidSchema(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	message1 := &TreeNode{
		Name:     "message_1",
		Type:     "message",
		FieldNum: 1,
		Children: make([]*TreeNode, 0),
	}

	child1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test1",
		Children: make([]*TreeNode, 0),
	}
	message1.AddChild(child1)

	child2 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	message1.AddChild(child2)

	root.AddChild(message1)

	message2 := &TreeNode{
		Name:     "message_1",
		Type:     "message",
		FieldNum: 2,
		Children: make([]*TreeNode, 0),
	}

	child3 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test2",
		Children: make([]*TreeNode, 0),
	}
	message2.AddChild(child3)

	child4 := &TreeNode{
		Name:     "field_2",
		Type:     "number",
		FieldNum: 2,
		Value:    "100",
		Children: make([]*TreeNode, 0),
	}
	message2.AddChild(child4)

	root.AddChild(message2)

	schema := serializer.GenerateProtoSchema(root)

	if strings.Count(schema, "message Message1 {") != 1 {
		t.Errorf("Expected exactly 1 definition of Message1, got %d occurrences. Schema:\n%s", strings.Count(schema, "message Message1 {"), schema)
	}

	if !strings.Contains(schema, "Message1 message_1 = 1;") {
		t.Errorf("Schema should contain Message1 message_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "Message1 message_1_2 = 2;") {
		t.Errorf("Schema should contain Message1 message_1_2 = 2, got:\n%s", schema)
	}

	messageDefCount := strings.Count(schema, "message Message1 {")
	if messageDefCount != 1 {
		t.Errorf("Message1 should be defined exactly once, found %d definitions", messageDefCount)
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestSerializeRaw_BoolFalseValuePreserved(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	serializer := NewSerializer(parser.GetProtocPath())

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	boolFieldFalse := &TreeNode{
		Name:     "field_1",
		Type:     "bool",
		FieldNum: 1,
		Value:    false,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(boolFieldFalse)

	boolFieldTrue := &TreeNode{
		Name:     "field_2",
		Type:     "bool",
		FieldNum: 2,
		Value:    true,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(boolFieldTrue)

	binaryData, err := serializer.SerializeRaw(root)
	if err != nil {
		t.Fatalf("Failed to serialize tree: %v", err)
	}

	if len(binaryData) == 0 {
		t.Fatal("Serialized data is empty")
	}

	parsedTree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse serialized data: %v", err)
	}

	if parsedTree == nil {
		t.Fatal("Parsed tree is nil")
	}

	var foundField1 bool
	var foundField2 bool
	var field1Value interface{}
	var field2Value interface{}

	for _, child := range parsedTree.Children {
		if child.FieldNum == 1 {
			foundField1 = true
			field1Value = child.Value
			if child.Type != "bool" {
				t.Errorf("Expected field_1 type to be 'bool', got '%s'", child.Type)
			}
		}
		if child.FieldNum == 2 {
			foundField2 = true
			field2Value = child.Value
			if child.Type != "bool" {
				t.Errorf("Expected field_2 type to be 'bool', got '%s'", child.Type)
			}
		}
	}

	if !foundField1 {
		t.Error("field_1 with bool value false was lost after serialization/deserialization")
	}

	if !foundField2 {
		t.Error("field_2 with bool value true was lost after serialization/deserialization")
	}

	if foundField1 {
		if field1Value != false {
			t.Errorf("Expected field_1 value to be false, got %v (type: %T)", field1Value, field1Value)
		}
	}

	if foundField2 {
		if field2Value != true {
			t.Errorf("Expected field_2 value to be true, got %v (type: %T)", field2Value, field2Value)
		}
	}
}