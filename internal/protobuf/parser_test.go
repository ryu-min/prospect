package protobuf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRawWithNestedMessage(t *testing.T) {
	testdataDir := "testdata"
	personBin := filepath.Join(testdataDir, "person.bin")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	if filepath.Base(wd) == "protobuf" {
		testdataDir = filepath.Join("..", "..", "testdata")
	}
	personBin = filepath.Join(testdataDir, "person.bin")

	absPath, err := filepath.Abs(personBin)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Skipping test: %s does not exist (tried %s from %s)", absPath, personBin, wd)
		return
	}

	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", absPath, err)
	}

	tree, err := parser.ParseRaw(data)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", absPath, err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	if tree.Name != "root" {
		t.Errorf("Expected root name to be 'root', got '%s'", tree.Name)
	}

	t.Logf("\nTree structure from ParseRaw:")
	printTree(t, tree, 0)

	expectedRootChildren := 6
	if len(tree.Children) != expectedRootChildren {
		t.Errorf("Expected root to have %d children, got %d", expectedRootChildren, len(tree.Children))
		t.Logf("Root children:")
		for i, child := range tree.Children {
			t.Logf("  [%d] %s (field_%d, %s) - children: %d", i, child.Name, child.FieldNum, child.Type, len(child.Children))
		}
	}

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

	if field4.Name != "field_4" {
		t.Errorf("Expected field_4 name to be 'field_4', got '%s'", field4.Name)
	}
	if !strings.HasPrefix(field4.Type, "message_") {
		t.Errorf("Expected field_4 to be of type starting with 'message_', got '%s'", field4.Type)
	}

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

	if field4Index != 3 {
		t.Errorf("Expected field_4 to be at index 3 in root.Children, got index %d", field4Index)
	}

	for i, child := range tree.Children {
		if i != field4Index && child.FieldNum >= 1 && child.FieldNum <= 4 {
			if i > field4Index {
				t.Errorf("Found field_%d at index %d (after field_4 at index %d) - this might indicate parsing issue", child.FieldNum, i, field4Index)
			}
		}
	}

	expectedFields := []struct {
		fieldNum int
		typeName string
		value    interface{}
	}{
		{1, "string", "123 Main St"},
		{2, "string", "New York alo"},
		{3, "string", "USA"},
		{4, "number", "12"},
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

func TestParseRaw_RepeatedFields(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "prospect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := `syntax = "proto3";

message TestMessage {
  repeated string field_1 = 1;
}
`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	textFormat := `field_1: "value1"
field_1: "value2"
field_1: "value3"
`

	encodeCmd := exec.Command(protocPath, "--encode", "TestMessage", "--proto_path", tempDir, protoFile)
	encodeCmd.Stdin = strings.NewReader(textFormat)
	binaryData, err := encodeCmd.Output()
	if err != nil {
		t.Fatalf("Failed to encode test data: %v", err)
	}

	tree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	field1Count := 0
	for _, child := range tree.Children {
		if child.FieldNum == 1 {
			field1Count++
		}
	}

	if field1Count < 3 {
		t.Errorf("Expected at least 3 instances of field_1, got %d", field1Count)
	}

	t.Logf("Parsed repeated fields: found %d instances of field_1", field1Count)
}

func TestParseRaw_NegativeNumber(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "prospect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := `syntax = "proto3";

message TestMessage {
  int64 field_1 = 1;
}
`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	textFormat := `field_1: -30
`

	encodeCmd := exec.Command(protocPath, "--encode", "TestMessage", "--proto_path", tempDir, protoFile)
	encodeCmd.Stdin = strings.NewReader(textFormat)
	binaryData, err := encodeCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatalf("Failed to encode test data: %v\nstderr: %s", err, string(exitError.Stderr))
		}
		t.Fatalf("Failed to encode test data: %v", err)
	}

	tree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	var field1 *TreeNode
	for _, child := range tree.Children {
		if child.FieldNum == 1 {
			field1 = child
			break
		}
	}

	if field1 == nil {
		t.Fatal("field_1 not found")
	}

	if field1.Type != "number" {
		t.Errorf("Expected field_1 type to be 'number', got '%s'", field1.Type)
	}

	expectedValue := "-30"
	if field1.Value != expectedValue {
		t.Errorf("Expected field_1 value to be '%s', got '%v'", expectedValue, field1.Value)
	}

	t.Logf("Successfully parsed negative number: field_1 = %v (type: %s)", field1.Value, field1.Type)
}

func TestParseRaw_MessageFieldNaming(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "prospect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := `syntax = "proto3";

message TestMessage {
  string field_1 = 1;
  NestedMessage field_4 = 4;
  int64 field_5 = 5;
  message NestedMessage {
    string field_1 = 1;
    int64 field_2 = 2;
  }
}
`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	textFormat := `field_1: "test"
field_4 {
  field_1: "nested"
  field_2: 42
}
field_5: 100
`

	encodeCmd := exec.Command(protocPath, "--encode", "TestMessage", "--proto_path", tempDir, protoFile)
	encodeCmd.Stdin = strings.NewReader(textFormat)
	binaryData, err := encodeCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatalf("Failed to encode test data: %v\nstderr: %s", err, string(exitError.Stderr))
		}
		t.Fatalf("Failed to encode test data: %v", err)
	}

	tree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	var field1 *TreeNode
	var field4 *TreeNode
	var field5 *TreeNode

	for _, child := range tree.Children {
		switch child.FieldNum {
		case 1:
			field1 = child
		case 4:
			field4 = child
		case 5:
			field5 = child
		}
	}

	if field1 == nil {
		t.Fatal("field_1 not found")
	}
	if field4 == nil {
		t.Fatal("field_4 (message) not found")
	}
	if field5 == nil {
		t.Fatal("field_5 not found")
	}

	if field1.Name != "field_1" {
		t.Errorf("Expected field_1 name to be 'field_1', got '%s'", field1.Name)
	}
	if field1.FieldNum != 1 {
		t.Errorf("Expected field_1 FieldNum to be 1, got %d", field1.FieldNum)
	}
	if field1.Type != "string" {
		t.Errorf("Expected field_1 type to be 'string', got '%s'", field1.Type)
	}

	if field4.Name != "field_4" {
		t.Errorf("Expected field_4 (message) name to be 'field_4', got '%s'", field4.Name)
	}
	if field4.FieldNum != 4 {
		t.Errorf("Expected field_4 FieldNum to be 4, got %d", field4.FieldNum)
	}
	if !strings.HasPrefix(field4.Type, "message_") {
		t.Errorf("Expected field_4 type to start with 'message_', got '%s'", field4.Type)
	}

	if field5.Name != "field_5" {
		t.Errorf("Expected field_5 name to be 'field_5', got '%s'", field5.Name)
	}
	if field5.FieldNum != 5 {
		t.Errorf("Expected field_5 FieldNum to be 5, got %d", field5.FieldNum)
	}
	if field5.Type != "number" {
		t.Errorf("Expected field_5 type to be 'number', got '%s'", field5.Type)
	}

	if len(field4.Children) < 2 {
		t.Errorf("Expected field_4 to have at least 2 children, got %d", len(field4.Children))
	}

	for _, child := range field4.Children {
		if child.FieldNum == 1 {
			if child.Name != "field_1" {
				t.Errorf("Expected nested field_1 name to be 'field_1', got '%s'", child.Name)
			}
			if child.Type != "string" {
				t.Errorf("Expected nested field_1 type to be 'string', got '%s'", child.Type)
			}
		}
		if child.FieldNum == 2 {
			if child.Name != "field_2" {
				t.Errorf("Expected nested field_2 name to be 'field_2', got '%s'", child.Name)
			}
			if child.Type != "number" {
				t.Errorf("Expected nested field_2 type to be 'number', got '%s'", child.Type)
			}
		}
	}

	t.Logf("Message field naming test passed:")
	t.Logf("  field_1: name=%s, fieldNum=%d, type=%s", field1.Name, field1.FieldNum, field1.Type)
	t.Logf("  field_4: name=%s, fieldNum=%d, type=%s", field4.Name, field4.FieldNum, field4.Type)
	t.Logf("  field_5: name=%s, fieldNum=%d, type=%s", field5.Name, field5.FieldNum, field5.Type)
}

func TestParseRaw_NestedMessageFieldNaming(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "prospect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := `syntax = "proto3";

message TestMessage {
  string field_1 = 1;
  OuterMessage field_2 = 2;
  message OuterMessage {
    string field_1 = 1;
    InnerMessage field_3 = 3;
    message InnerMessage {
      string field_1 = 1;
      int64 field_2 = 2;
    }
  }
}
`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	textFormat := `field_1: "root"
field_2 {
  field_1: "outer"
  field_3 {
    field_1: "inner"
    field_2: 42
  }
}
`

	encodeCmd := exec.Command(protocPath, "--encode", "TestMessage", "--proto_path", tempDir, protoFile)
	encodeCmd.Stdin = strings.NewReader(textFormat)
	binaryData, err := encodeCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatalf("Failed to encode test data: %v\nstderr: %s", err, string(exitError.Stderr))
		}
		t.Fatalf("Failed to encode test data: %v", err)
	}

	tree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	var field2 *TreeNode
	for _, child := range tree.Children {
		if child.FieldNum == 2 {
			field2 = child
			break
		}
	}

	if field2 == nil {
		t.Fatal("field_2 (OuterMessage) not found")
	}

	if field2.Name != "field_2" {
		t.Errorf("Expected field_2 name to be 'field_2', got '%s'", field2.Name)
	}
	if field2.FieldNum != 2 {
		t.Errorf("Expected field_2 FieldNum to be 2, got %d", field2.FieldNum)
	}
	if !strings.HasPrefix(field2.Type, "message_") {
		t.Errorf("Expected field_2 type to start with 'message_', got '%s'", field2.Type)
	}

	var field3 *TreeNode
	for _, child := range field2.Children {
		if child.FieldNum == 3 {
			field3 = child
			break
		}
	}

	if field3 == nil {
		t.Fatal("field_3 (InnerMessage) not found in field_2")
	}

	if field3.Name != "field_3" {
		t.Errorf("Expected nested field_3 name to be 'field_3', got '%s'", field3.Name)
	}
	if field3.FieldNum != 3 {
		t.Errorf("Expected nested field_3 FieldNum to be 3, got %d", field3.FieldNum)
	}
	if !strings.HasPrefix(field3.Type, "message_") {
		t.Errorf("Expected nested field_3 type to start with 'message_', got '%s'", field3.Type)
	}

	if field2.Type == field3.Type {
		t.Errorf("Expected field_2 and field_3 to have different message types, both got '%s'", field2.Type)
	}

	t.Logf("Nested message field naming test passed:")
	t.Logf("  field_2: name=%s, fieldNum=%d, type=%s", field2.Name, field2.FieldNum, field2.Type)
	t.Logf("  field_3: name=%s, fieldNum=%d, type=%s", field3.Name, field3.FieldNum, field3.Type)
}

func TestParseRaw_MessageFieldNameMatchesFieldNum(t *testing.T) {
	protocPath := "protoc"
	if _, err := exec.LookPath(protocPath); err != nil {
		t.Skipf("Skipping test: protoc not found in PATH")
		return
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "prospect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := `syntax = "proto3";

message TestMessage {
  string field_1 = 1;
  NestedMessage field_4 = 4;
  int64 field_5 = 5;
  NestedMessage field_7 = 7;
  message NestedMessage {
    string field_1 = 1;
    int64 field_2 = 2;
  }
}
`
	protoFile := filepath.Join(tempDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		t.Fatalf("Failed to write proto file: %v", err)
	}

	textFormat := `field_1: "test"
field_4 {
  field_1: "nested1"
  field_2: 10
}
field_5: 100
field_7 {
  field_1: "nested2"
  field_2: 20
}
`

	encodeCmd := exec.Command(protocPath, "--encode", "TestMessage", "--proto_path", tempDir, protoFile)
	encodeCmd.Stdin = strings.NewReader(textFormat)
	binaryData, err := encodeCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatalf("Failed to encode test data: %v\nstderr: %s", err, string(exitError.Stderr))
		}
		t.Fatalf("Failed to encode test data: %v", err)
	}

	tree, err := parser.ParseRaw(binaryData)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if tree == nil {
		t.Fatal("Tree is nil")
	}

	checkFieldName := func(node *TreeNode, expectedFieldNum int) {
		expectedName := fmt.Sprintf("field_%d", expectedFieldNum)
		if node.Name != expectedName {
			t.Errorf("Expected field with FieldNum=%d to have name '%s', got '%s'", expectedFieldNum, expectedName, node.Name)
		}
		if node.FieldNum != expectedFieldNum {
			t.Errorf("Expected field name '%s' to have FieldNum=%d, got %d", expectedName, expectedFieldNum, node.FieldNum)
		}
	}

	for _, child := range tree.Children {
		checkFieldName(child, child.FieldNum)

		if strings.HasPrefix(child.Type, "message_") {
			if child.Name != fmt.Sprintf("field_%d", child.FieldNum) {
				t.Errorf("Message field with FieldNum=%d should have name 'field_%d', got '%s'", child.FieldNum, child.FieldNum, child.Name)
			}

			for _, nestedChild := range child.Children {
				checkFieldName(nestedChild, nestedChild.FieldNum)
			}
		}
	}

	t.Logf("All field names match their FieldNum values")
}
