package ui

import (
	"testing"

	"prospect/internal/protobuf"
)

func TestHandleTypeChangeToMessageWithExistingChildren(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	existingChild1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "string",
		Value:    "test",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	existingChild2 := &protobuf.TreeNode{
		Name:     "field_2",
		Type:     "number",
		Value:    "42",
		FieldNum: 2,
		Children: make([]*protobuf.TreeNode, 0),
	}

	nodeToChange := &protobuf.TreeNode{
		Name:     "field_3",
		Type:     "string",
		Value:    "old_value",
		FieldNum: 3,
		Children: []*protobuf.TreeNode{existingChild1, existingChild2},
	}

	root.Children = append(root.Children, nodeToChange)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("0", "string", "message_1")

	if nodeToChange.Type != "message_1" {
		t.Errorf("Expected type to be 'message_1', got '%s'", nodeToChange.Type)
	}

	if len(nodeToChange.Children) != 2 {
		t.Errorf("Expected 2 children to be preserved, got %d", len(nodeToChange.Children))
	}

	if nodeToChange.Children[0].FieldNum != 1 || nodeToChange.Children[1].FieldNum != 2 {
		t.Errorf("Expected children to be preserved with field numbers 1 and 2")
	}

	if nodeToChange.Value != nil {
		t.Errorf("Expected Value to be nil, got %v", nodeToChange.Value)
	}
}

func TestHandleTypeChangeToMessageWithoutChildrenWithSourceMessage(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	sourceMessage := &protobuf.TreeNode{
		Name:     "field_10",
		Type:     "message_1",
		FieldNum: 10,
		Children: []*protobuf.TreeNode{
			{
				Name:     "field_1",
				Type:     "string",
				Value:    "source_value",
				FieldNum: 1,
				Children: make([]*protobuf.TreeNode, 0),
			},
			{
				Name:     "field_2",
				Type:     "number",
				Value:    "100",
				FieldNum: 2,
				Children: make([]*protobuf.TreeNode, 0),
			},
		},
	}

	nodeToChange := &protobuf.TreeNode{
		Name:     "field_3",
		Type:     "string",
		Value:    "old_value",
		FieldNum: 3,
		Children: make([]*protobuf.TreeNode, 0),
	}

	root.Children = append(root.Children, sourceMessage, nodeToChange)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("1", "string", "message_1")

	if nodeToChange.Type != "message_1" {
		t.Errorf("Expected type to be 'message_1', got '%s'", nodeToChange.Type)
	}

	if len(nodeToChange.Children) != 2 {
		t.Errorf("Expected 2 children to be copied from source message, got %d", len(nodeToChange.Children))
	}

	if nodeToChange.Children[0].FieldNum != 1 || nodeToChange.Children[1].FieldNum != 2 {
		t.Errorf("Expected children to be copied with field numbers 1 and 2")
	}

	if nodeToChange.Children[0].Value != "source_value" {
		t.Errorf("Expected first child value to be 'source_value', got %v", nodeToChange.Children[0].Value)
	}

	if nodeToChange.Value != nil {
		t.Errorf("Expected Value to be nil, got %v", nodeToChange.Value)
	}
}

func TestHandleTypeChangeToMessageWithoutChildrenWithoutSourceMessage(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	nodeToChange := &protobuf.TreeNode{
		Name:     "field_3",
		Type:     "string",
		Value:    "old_value",
		FieldNum: 3,
		Children: make([]*protobuf.TreeNode, 0),
	}

	root.Children = append(root.Children, nodeToChange)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("0", "string", "message_999")

	if !adapter.isMessageType(nodeToChange.Type) {
		t.Errorf("Expected type to be a message type, got '%s'", nodeToChange.Type)
	}

	if len(nodeToChange.Children) != 0 {
		t.Errorf("Expected 0 children for new message, got %d", len(nodeToChange.Children))
	}

	if nodeToChange.Value != nil {
		t.Errorf("Expected Value to be nil, got %v", nodeToChange.Value)
	}
}

func TestHandleTypeChangeFromMessageToMessageWithExistingChildren(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	existingChild := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "string",
		Value:    "preserved",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	nodeToChange := &protobuf.TreeNode{
		Name:     "field_3",
		Type:     "message_1",
		FieldNum: 3,
		Children: []*protobuf.TreeNode{existingChild},
	}

	root.Children = append(root.Children, nodeToChange)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("0", "message_1", "message_2")

	if nodeToChange.Type != "message_2" {
		t.Errorf("Expected type to be 'message_2', got '%s'", nodeToChange.Type)
	}

	if len(nodeToChange.Children) != 1 {
		t.Errorf("Expected 1 child to be preserved, got %d", len(nodeToChange.Children))
	}

	if nodeToChange.Children[0].Value != "preserved" {
		t.Errorf("Expected child value to be 'preserved', got %v", nodeToChange.Children[0].Value)
	}
}

func TestHandleTypeChangeFromMessageToMessageWithoutChildren(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	sourceMessage := &protobuf.TreeNode{
		Name:     "field_10",
		Type:     "message_2",
		FieldNum: 10,
		Children: []*protobuf.TreeNode{
			{
				Name:     "field_1",
				Type:     "string",
				Value:    "source",
				FieldNum: 1,
				Children: make([]*protobuf.TreeNode, 0),
			},
		},
	}

	nodeToChange := &protobuf.TreeNode{
		Name:     "field_3",
		Type:     "message_1",
		FieldNum: 3,
		Children: make([]*protobuf.TreeNode, 0),
	}

	root.Children = append(root.Children, sourceMessage, nodeToChange)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("1", "message_1", "message_2")

	if nodeToChange.Type != "message_2" {
		t.Errorf("Expected type to be 'message_2', got '%s'", nodeToChange.Type)
	}

	if len(nodeToChange.Children) != 1 {
		t.Errorf("Expected 1 child to be copied from source message, got %d", len(nodeToChange.Children))
	}

	if nodeToChange.Children[0].Value != "source" {
		t.Errorf("Expected child value to be 'source', got %v", nodeToChange.Children[0].Value)
	}
}

