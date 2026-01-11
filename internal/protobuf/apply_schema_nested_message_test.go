package protobuf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplySchema_NestedMessageType(t *testing.T) {
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

	// Создаем поле с типом message (которое должно стать Message1)
	messageField := &TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	messageField.AddChild(nestedField1)

	root.AddChild(messageField)

	// Создаем еще одно поле с типом message (field_3)
	messageField2 := &TreeNode{
		Name:     "field_3",
		Type:     "message_2",
		FieldNum: 3,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(messageField2)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional Message1 first_field = 1;
  message Message1 {
    optional string field_1 = 1;
    optional string field_2 = 2;
  }
  optional Message1 field_3 = 3;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Проверяем, что поле field_1 имеет тип Message1
	foundFirstField := false
	foundField3 := false
	for _, child := range root.Children {
		if child.FieldNum == 1 {
			if child.Name != "first_field" {
				t.Errorf("Expected field name to be 'first_field', got '%s'", child.Name)
			}
			if child.Type != "Message1" {
				t.Errorf("Expected field type to be 'Message1', got '%s'", child.Type)
			}
			foundFirstField = true
		} else if child.FieldNum == 3 {
			if child.Name != "field_3" {
				t.Errorf("Expected field name to be 'field_3', got '%s'", child.Name)
			}
			if child.Type != "Message1" {
				t.Errorf("Expected field type to be 'Message1', got '%s'", child.Type)
			}
			foundField3 = true
		}
	}

	if !foundFirstField {
		t.Error("Field 'first_field' not found")
	}
	if !foundField3 {
		t.Error("Field 'field_3' not found")
	}
}

func TestApplySchema_NestedMessageTypeWithSchema1(t *testing.T) {
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

	// Создаем поле с типом message (которое должно стать Message1)
	messageField := &TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*TreeNode, 0),
	}

	nestedField1 := &TreeNode{
		Name:     "field_1",
		Type:     "string",
		FieldNum: 1,
		Value:    "test",
		Children: make([]*TreeNode, 0),
	}
	messageField.AddChild(nestedField1)

	root.AddChild(messageField)

	// Создаем еще одно поле с типом message (field_3)
	messageField2 := &TreeNode{
		Name:     "field_3",
		Type:     "message_2",
		FieldNum: 3,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(messageField2)

	// Создаем еще одно поле с типом message (field_4)
	messageField3 := &TreeNode{
		Name:     "field_4",
		Type:     "message_3",
		FieldNum: 4,
		Children: make([]*TreeNode, 0),
	}
	root.AddChild(messageField3)

	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "test.proto")
	schemaContent := `syntax = "proto2";

message Message {
  optional Message1 first_field = 1;
  message Message1 {
    optional string field_1 = 1;
    optional string field_2 = 2;
    optional int64 field_3 = 3;
    optional bool field_4 = 4;
  }
  optional Message1 field_3 = 3;
  optional Message1 field_4 = 4;
  repeated string field_5 = 5;
}`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	_, err = parser.ApplySchema(root, schemaFile)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Проверяем, что все поля с типом Message1 имеют правильный тип
	for _, child := range root.Children {
		if child.FieldNum == 1 || child.FieldNum == 3 || child.FieldNum == 4 {
			if child.Type != "Message1" {
				t.Errorf("Field %d: Expected type to be 'Message1', got '%s'", child.FieldNum, child.Type)
			}
		}
	}
}

