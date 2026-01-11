package protobuf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplySchema_SimpleFields(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

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
		Value:    "test_value",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "int64",
		FieldNum: 2,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional string name = 1;
  optional int32 age = 2;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(root.Children))
	}

	foundName := false
	foundAge := false
	for _, child := range root.Children {
		if child.FieldNum == 1 {
			if child.Name != "name" {
				t.Errorf("Expected field name to be 'name', got '%s'", child.Name)
			}
			if child.Type != "string" {
				t.Errorf("Expected field type to be 'string', got '%s'", child.Type)
			}
			foundName = true
		} else if child.FieldNum == 2 {
			if child.Name != "age" {
				t.Errorf("Expected field name to be 'age', got '%s'", child.Name)
			}
			if child.Type != "int32" {
				t.Errorf("Expected field type to be 'int32', got '%s'", child.Type)
			}
			foundAge = true
		}
	}

	if !foundName {
		t.Error("Field 'name' not found")
	}
	if !foundAge {
		t.Error("Field 'age' not found")
	}
}

func TestApplySchema_TypeConversion(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:     "field_1",
		Type:     "int64",
		FieldNum: 1,
		Value:    "42",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional int32 number = 1;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.Name != "number" {
		t.Errorf("Expected field name to be 'number', got '%s'", child.Name)
	}
	if child.Type != "int32" {
		t.Errorf("Expected field type to be 'int32', got '%s'", child.Type)
	}
}

func TestApplySchema_RequiredFieldMissing(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

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

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional string name = 1;
  required int32 age = 2;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err == nil {
		t.Fatal("Expected error for missing required field, but got none")
	}

	if !strings.Contains(err.Error(), "отсутствует обязательное поле") {
		t.Errorf("Expected error message about required field, got: %v", err)
	}
}

func TestApplySchema_NestedMessage(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	nestedMessage := &TreeNode{
		Name:     "field_4",
		Type:     "message_1",
		FieldNum: 4,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "street_value",
		Children: make([]*TreeNode, 0),
	}
	nestedMessage.AddChild(nestedField1)

	root.AddChild(nestedMessage)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional Address address = 4;
}

message Address {
  optional string street = 1;
  optional string city = 2;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.Name != "address" {
		t.Errorf("Expected field name to be 'address', got '%s'", child.Name)
	}

	if len(child.Children) != 1 {
		t.Fatalf("Expected 1 nested child, got %d", len(child.Children))
	}

	nestedChild := child.Children[0]
	if nestedChild.Name != "street" {
		t.Errorf("Expected nested field name to be 'street', got '%s'", nestedChild.Name)
	}
}

func TestApplySchema_RepeatedField(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		FieldNum: 0,
		Children: make([]*TreeNode, 0),
	}

	field1 := &TreeNode{
		Name:       "field_5",
		Type:       "string",
		FieldNum:   5,
		Value:      "hobby1",
		IsRepeated: false,
		Children:   make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  repeated string hobbies = 5;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.Name != "hobbies" {
		t.Errorf("Expected field name to be 'hobbies', got '%s'", child.Name)
	}
	if !child.IsRepeated {
		t.Error("Expected field to be repeated")
	}
}

func TestApplySchema_Proto3Syntax(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

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
		Value:    "John",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field1)

	field2 := &TreeNode{
		Name:     "field_2",
		Type:     "int64",
		FieldNum: 2,
		Value:    "30",
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(field2)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto3";

message Message {
  string name = 1;
  int32 age = 2;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	foundName := false
	foundAge := false
	for _, child := range root.Children {
		if child.FieldNum == 1 {
			if child.Name != "name" {
				t.Errorf("Expected field name to be 'name', got '%s'", child.Name)
			}
			foundName = true
		} else if child.FieldNum == 2 {
			if child.Name != "age" {
				t.Errorf("Expected field name to be 'age', got '%s'", child.Name)
			}
			if child.Type != "int32" {
				t.Errorf("Expected field type to be 'int32', got '%s'", child.Type)
			}
			foundAge = true
		}
	}

	if !foundName {
		t.Error("Field 'name' not found")
	}
	if !foundAge {
		t.Error("Field 'age' not found")
	}
}

