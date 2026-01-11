package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportProtoSchema_AllFieldTypes(t *testing.T) {
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

	fields := []struct {
		name     string
		fieldType string
		fieldNum  int
		value     interface{}
	}{
		{"field_1", "string", 1, "test"},
		{"field_2", "int32", 2, "42"},
		{"field_3", "int64", 3, "123456789"},
		{"field_4", "uint32", 4, "100"},
		{"field_5", "uint64", 5, "200"},
		{"field_6", "sint32", 6, "-10"},
		{"field_7", "sint64", 7, "-20"},
		{"field_8", "bool", 8, true},
		{"field_9", "float", 9, "3.14"},
		{"field_10", "double", 10, "2.718"},
	}

	for _, f := range fields {
		field := &TreeNode{
			Name:     f.name,
			Type:     f.fieldType,
			FieldNum: f.fieldNum,
			Value:    f.value,
			Children: make([]*TreeNode, 0),
		}
		root.AddChild(field)
	}

	schema := serializer.GenerateProtoSchema(root)

	if !strings.Contains(schema, "syntax = \"proto2\";") {
		t.Error("Schema should contain syntax declaration")
	}

	if !strings.Contains(schema, "message Message {") {
		t.Error("Schema should contain Message declaration")
	}

	for _, f := range fields {
		expectedField := fmt.Sprintf("optional %s %s = %d;", f.fieldType, f.name, f.fieldNum)
		if !strings.Contains(schema, expectedField) {
			t.Errorf("Schema should contain %s, got:\n%s", expectedField, schema)
		}
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestExportProtoSchema_RepeatedFields(t *testing.T) {
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

	field1 := &TreeNode{
		Name:       "field_1",
		Type:       "string",
		FieldNum:   1,
		Value:      "test1",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:       "field_2",
		Type:       "int32",
		FieldNum:   2,
		Value:      "42",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	schema := serializer.GenerateProtoSchema(root)

	if !strings.Contains(schema, "repeated string field_1 = 1;") {
		t.Errorf("Schema should contain repeated string field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "repeated int32 field_2 = 2;") {
		t.Errorf("Schema should contain repeated int32 field_2 = 2, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestExportProtoSchema_NestedMessages(t *testing.T) {
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

	nestedMessage := &TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "nested_field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "nested_value",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField1)

	nestedField2 := &TreeNode{
		Name:     "nested_field_2",
		Type:     "int32",
		FieldNum: 2,
		Value:    "100",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField2)

	root.AddChild(nestedMessage)

	schema := serializer.GenerateProtoSchema(root)

	if !strings.Contains(schema, "optional Message1 field_1 = 1;") {
		t.Errorf("Schema should contain optional Message1 field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "message Message1 {") {
		t.Errorf("Schema should contain message Message1 declaration, got:\n%s", schema)
	}

	if !strings.Contains(schema, "optional string nested_field_1 = 1;") {
		t.Errorf("Schema should contain optional string nested_field_1 = 1, got:\n%s", schema)
	}

	if !strings.Contains(schema, "optional int32 nested_field_2 = 2;") {
		t.Errorf("Schema should contain optional int32 nested_field_2 = 2, got:\n%s", schema)
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestExportProtoSchema_ComplexStructure(t *testing.T) {
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

	simpleField := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "simple",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(simpleField)

	repeatedField := &TreeNode{
		Name:       "field_2",
		Type:       "int32",
		FieldNum:   2,
		Value:      "42",
		IsRepeated: true,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(repeatedField)

	nestedMessage := &TreeNode{
		Name:     "field_3",
		Type:     "message_1",
		FieldNum: 3,
		Children: make([]*TreeNode, 0),
	}

	nestedField := &TreeNode{
		Name:     "nested_field",
		Type:     "double",
		FieldNum: 1,
		Value:    "3.14",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField)

	root.AddChild(nestedMessage)

	schema := serializer.GenerateProtoSchema(root)

	if !strings.Contains(schema, "optional string field_1 = 1;") {
		t.Errorf("Schema should contain optional string field_1 = 1")
	}

	if !strings.Contains(schema, "repeated int32 field_2 = 2;") {
		t.Errorf("Schema should contain repeated int32 field_2 = 2")
	}

	if !strings.Contains(schema, "optional Message1 field_3 = 3;") {
		t.Errorf("Schema should contain optional Message1 field_3 = 3")
	}

	if !strings.Contains(schema, "message Message1 {") {
		t.Errorf("Schema should contain message Message1 declaration")
	}

	if !strings.Contains(schema, "optional double nested_field = 1;") {
		t.Errorf("Schema should contain optional double nested_field = 1")
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestExportProtoSchema_EmptyTree(t *testing.T) {
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

	schema := serializer.GenerateProtoSchema(root)

	if !strings.Contains(schema, "syntax = \"proto2\";") {
		t.Error("Schema should contain syntax declaration")
	}

	if !strings.Contains(schema, "message Message {") {
		t.Error("Schema should contain Message declaration")
	}

	if !strings.Contains(schema, "}") {
		t.Error("Schema should contain closing brace")
	}

	t.Logf("Generated schema:\n%s", schema)
}

func TestExportProtoSchema_FileOutput(t *testing.T) {
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
		Type:     "int32",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	schema := serializer.GenerateProtoSchema(root)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test_schema.proto")

	err = os.WriteFile(schemaFile, []byte(schema), 0644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	readSchema, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	if string(readSchema) != schema {
		t.Error("Written schema does not match generated schema")
	}

	if !strings.Contains(string(readSchema), "syntax = \"proto2\";") {
		t.Error("Written schema should contain syntax declaration")
	}

	if !strings.Contains(string(readSchema), "message Message {") {
		t.Error("Written schema should contain Message declaration")
	}

	t.Logf("Schema written to file successfully:\n%s", string(readSchema))
}

