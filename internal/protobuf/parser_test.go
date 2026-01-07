package protobuf

import (
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

	if field4.Type != "message" {
		t.Errorf("Expected field_4 to be of type 'message', got '%s'", field4.Type)
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
