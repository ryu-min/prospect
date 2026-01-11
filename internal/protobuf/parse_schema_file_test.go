package protobuf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSchemaFile_SingleMessage(t *testing.T) {
	// Создаем временный файл с одним сообщением
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.proto")
	schemaContent := `syntax = "proto2";

message Person {
  optional string name = 1;
  optional int32 age = 2;
}
`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	messageNames, err := parser.ParseSchemaFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to parse schema file: %v", err)
	}

	if len(messageNames) != 1 {
		t.Fatalf("Expected 1 message, got %d: %v", len(messageNames), messageNames)
	}

	if messageNames[0] != "Person" {
		t.Fatalf("Expected message name 'Person', got '%s'", messageNames[0])
	}
}

func TestParseSchemaFile_MultipleMessages(t *testing.T) {
	// Создаем временный файл с несколькими сообщениями
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.proto")
	schemaContent := `syntax = "proto2";

message FirstMessage {
  optional string field1 = 1;
}

message SecondMessage {
  optional int32 field1 = 1;
}

message ThirdMessage {
  optional bool field1 = 1;
}
`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	messageNames, err := parser.ParseSchemaFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to parse schema file: %v", err)
	}

	if len(messageNames) != 3 {
		t.Fatalf("Expected 3 messages, got %d: %v", len(messageNames), messageNames)
	}

	expectedNames := map[string]bool{
		"FirstMessage":  true,
		"SecondMessage": true,
		"ThirdMessage":  true,
	}

	for _, name := range messageNames {
		if !expectedNames[name] {
			t.Errorf("Unexpected message name: %s", name)
		}
	}
}

func TestParseSchemaFile_WithNestedMessages(t *testing.T) {
	// Создаем временный файл с вложенными сообщениями
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.proto")
	schemaContent := `syntax = "proto2";

message OuterMessage {
  optional string field1 = 1;
  message NestedMessage {
    optional int32 field1 = 1;
  }
}

message AnotherMessage {
  optional bool field1 = 1;
}
`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	messageNames, err := parser.ParseSchemaFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to parse schema file: %v", err)
	}

	// Должны быть только сообщения верхнего уровня
	if len(messageNames) != 2 {
		t.Fatalf("Expected 2 top-level messages, got %d: %v", len(messageNames), messageNames)
	}

	expectedNames := map[string]bool{
		"OuterMessage":  true,
		"AnotherMessage": true,
	}

	for _, name := range messageNames {
		if !expectedNames[name] {
			t.Errorf("Unexpected message name: %s", name)
		}
	}
}

