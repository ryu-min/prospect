package protobuf

import (
	"encoding/json"
	"testing"
)

func TestTreeNodeToJSON_SimpleFields(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{Name: "name", Type: "string", Value: "test", FieldNum: 1},
			{Name: "age", Type: "int64", Value: "25", FieldNum: 2},
			{Name: "active", Type: "bool", Value: true, FieldNum: 3},
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name to be 'test', got %v", result["name"])
	}
	if result["age"] != "25" {
		t.Errorf("Expected age to be '25', got %v", result["age"])
	}
	if result["active"] != true {
		t.Errorf("Expected active to be true, got %v", result["active"])
	}
}

func TestTreeNodeToJSON_NestedMessage(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{
				Name:     "user",
				Type:     "message",
				FieldNum: 1,
				Children: []*TreeNode{
					{Name: "name", Type: "string", Value: "John", FieldNum: 1},
					{Name: "email", Type: "string", Value: "john@example.com", FieldNum: 2},
				},
			},
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected user to be a map, got %T", result["user"])
	}

	if user["name"] != "John" {
		t.Errorf("Expected user.name to be 'John', got %v", user["name"])
	}
	if user["email"] != "john@example.com" {
		t.Errorf("Expected user.email to be 'john@example.com', got %v", user["email"])
	}
}

func TestTreeNodeToJSON_RepeatedField(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{Name: "tags", Type: "string", Value: "tag1", FieldNum: 1, IsRepeated: true},
			{Name: "tags", Type: "string", Value: "tag2", FieldNum: 1, IsRepeated: true},
			{Name: "tags", Type: "string", Value: "tag3", FieldNum: 1, IsRepeated: true},
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatalf("Expected tags to be an array, got %T", result["tags"])
	}

	if len(tags) != 3 {
		t.Errorf("Expected tags to have 3 elements, got %d", len(tags))
	}

	expectedTags := []string{"tag1", "tag2", "tag3"}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tags[%d] to be '%s', got %v", i, expectedTags[i], tag)
		}
	}
}

func TestTreeNodeToJSON_RepeatedMessage(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{
				Name:       "items",
				Type:       "message",
				FieldNum:   1,
				IsRepeated: true,
				Children: []*TreeNode{
					{Name: "id", Type: "int64", Value: "1", FieldNum: 1},
					{Name: "name", Type: "string", Value: "Item1", FieldNum: 2},
				},
			},
			{
				Name:       "items",
				Type:       "message",
				FieldNum:   1,
				IsRepeated: true,
				Children: []*TreeNode{
					{Name: "id", Type: "int64", Value: "2", FieldNum: 1},
					{Name: "name", Type: "string", Value: "Item2", FieldNum: 2},
				},
			},
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatalf("Expected items to be an array, got %T", result["items"])
	}

	if len(items) != 2 {
		t.Errorf("Expected items to have 2 elements, got %d", len(items))
	}

	item1, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected items[0] to be a map, got %T", items[0])
	}
	if item1["id"] != "1" {
		t.Errorf("Expected items[0].id to be '1', got %v", item1["id"])
	}
	if item1["name"] != "Item1" {
		t.Errorf("Expected items[0].name to be 'Item1', got %v", item1["name"])
	}

	item2, ok := items[1].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected items[1] to be a map, got %T", items[1])
	}
	if item2["id"] != "2" {
		t.Errorf("Expected items[1].id to be '2', got %v", item2["id"])
	}
	if item2["name"] != "Item2" {
		t.Errorf("Expected items[1].name to be 'Item2', got %v", item2["name"])
	}
}

func TestTreeNodeToJSON_FieldNameFromSchema(t *testing.T) {
	// Тест проверяет, что используется Name из ноды (которое содержит имя из схемы если она была применена)
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{Name: "user_name", Type: "string", Value: "John", FieldNum: 1},
			{Name: "", Type: "string", Value: "Unknown", FieldNum: 2}, // Если Name пустое, используется field_N
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	if result["user_name"] != "John" {
		t.Errorf("Expected user_name to be 'John', got %v", result["user_name"])
	}

	if result["field_2"] != "Unknown" {
		t.Errorf("Expected field_2 to be 'Unknown', got %v", result["field_2"])
	}
}

func TestTreeNodeToJSON_DeeplyNested(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{
				Name:     "company",
				Type:     "message",
				FieldNum: 1,
				Children: []*TreeNode{
					{
						Name:     "address",
						Type:     "message",
						FieldNum: 1,
						Children: []*TreeNode{
							{Name: "street", Type: "string", Value: "123 Main St", FieldNum: 1},
							{Name: "city", Type: "string", Value: "New York", FieldNum: 2},
						},
					},
				},
			},
		},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	company, ok := result["company"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected company to be a map, got %T", result["company"])
	}

	address, ok := company["address"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected company.address to be a map, got %T", company["address"])
	}

	if address["street"] != "123 Main St" {
		t.Errorf("Expected company.address.street to be '123 Main St', got %v", address["street"])
	}
	if address["city"] != "New York" {
		t.Errorf("Expected company.address.city to be 'New York', got %v", address["city"])
	}
}

func TestTreeNodeToJSONString(t *testing.T) {
	root := &TreeNode{
		Name: "root",
		Type: "message",
		Children: []*TreeNode{
			{Name: "name", Type: "string", Value: "test", FieldNum: 1},
			{Name: "age", Type: "int64", Value: "25", FieldNum: 2},
		},
	}

	jsonStr, err := TreeNodeToJSONString(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSONString failed: %v", err)
	}

	// Проверяем, что это валидный JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Generated JSON is invalid: %v\nJSON: %s", err, jsonStr)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name to be 'test', got %v", result["name"])
	}
	if result["age"] != "25" {
		t.Errorf("Expected age to be '25', got %v", result["age"])
	}
}

func TestTreeNodeToJSON_NilNode(t *testing.T) {
	_, err := TreeNodeToJSON(nil)
	if err == nil {
		t.Error("Expected error for nil node, got nil")
	}
}

func TestTreeNodeToJSON_EmptyRoot(t *testing.T) {
	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		Children: []*TreeNode{},
	}

	result, err := TreeNodeToJSON(root)
	if err != nil {
		t.Fatalf("TreeNodeToJSON failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty map for root with no children, got %v", result)
	}
}

