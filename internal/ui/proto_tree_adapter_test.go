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

func TestFindParentMessage(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "string",
		Value:    "test",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1.Children = append(message1.Children, field1)
	root.Children = append(root.Children, message1)

	adapter := newProtoTreeAdapter(root)

	parent := adapter.findParentMessage(field1)
	if parent == nil {
		t.Fatal("Expected to find parent message, got nil")
	}

	if parent != message1 {
		t.Errorf("Expected parent to be message1, got %v", parent)
	}

	if parent.Type != "message_1" {
		t.Errorf("Expected parent type to be 'message_1', got '%s'", parent.Type)
	}
}

func TestFindFieldsWithSameFieldNumInMessageType(t *testing.T) {
	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message2 := &protobuf.TreeNode{
		Name:     "field_2",
		Type:     "message_1",
		FieldNum: 2,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field1InMessage1 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "string",
		Value:    "value1",
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field1InMessage2 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "string",
		Value:    "value2",
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1.Children = append(message1.Children, field1InMessage1)
	message2.Children = append(message2.Children, field1InMessage2)
	root.Children = append(root.Children, message1, message2)

	adapter := newProtoTreeAdapter(root)

	affectedFields := adapter.findFieldsWithSameFieldNumInMessageType(field1InMessage1, "message_1", 5)
	if len(affectedFields) != 1 {
		t.Errorf("Expected 1 affected field, got %d", len(affectedFields))
	}

	if affectedFields[0] != field1InMessage2 {
		t.Errorf("Expected affected field to be field1InMessage2, got %v", affectedFields[0])
	}
}

func TestHandleTypeChangeWithFieldTypeSync(t *testing.T) {
	dontAskFieldTypeSyncConfirmation = true
	defer func() {
		dontAskFieldTypeSyncConfirmation = false
	}()

	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message2 := &protobuf.TreeNode{
		Name:     "field_2",
		Type:     "message_1",
		FieldNum: 2,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field5InMessage1 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "string",
		Value:    "value1",
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field5InMessage2 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "string",
		Value:    "value2",
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1.Children = append(message1.Children, field5InMessage1)
	message2.Children = append(message2.Children, field5InMessage2)
	root.Children = append(root.Children, message1, message2)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("0:0", "string", "number")

	if field5InMessage1.Type != "number" {
		t.Errorf("Expected field5InMessage1 type to be 'number', got '%s'", field5InMessage1.Type)
	}

	if field5InMessage2.Type != "number" {
		t.Errorf("Expected field5InMessage2 type to be 'number', got '%s'", field5InMessage2.Type)
	}

	if field5InMessage1.Value != nil {
		t.Errorf("Expected field5InMessage1 value to be nil, got %v", field5InMessage1.Value)
	}

	if field5InMessage2.Value != nil {
		t.Errorf("Expected field5InMessage2 value to be nil, got %v", field5InMessage2.Value)
	}
}

func TestHandleTypeChangeWithSeamlessChangeAndFieldTypeSync(t *testing.T) {
	dontAskFieldTypeSyncConfirmation = true
	defer func() {
		dontAskFieldTypeSyncConfirmation = false
	}()

	root := &protobuf.TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1 := &protobuf.TreeNode{
		Name:     "field_1",
		Type:     "message_1",
		FieldNum: 1,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message2 := &protobuf.TreeNode{
		Name:     "field_2",
		Type:     "message_1",
		FieldNum: 2,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field5InMessage1 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "bool",
		Value:    true,
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	field5InMessage2 := &protobuf.TreeNode{
		Name:     "field_5",
		Type:     "bool",
		Value:    false,
		FieldNum: 5,
		Children: make([]*protobuf.TreeNode, 0),
	}

	message1.Children = append(message1.Children, field5InMessage1)
	message2.Children = append(message2.Children, field5InMessage2)
	root.Children = append(root.Children, message1, message2)

	adapter := newProtoTreeAdapter(root)

	adapter.handleTypeChange("0:0", "bool", "number")

	if field5InMessage1.Type != "number" {
		t.Errorf("Expected field5InMessage1 type to be 'number', got '%s'", field5InMessage1.Type)
	}

	if field5InMessage2.Type != "number" {
		t.Errorf("Expected field5InMessage2 type to be 'number', got '%s'", field5InMessage2.Type)
	}

	if field5InMessage1.Value != "1" {
		t.Errorf("Expected field5InMessage1 value to be '1', got %v", field5InMessage1.Value)
	}

	if field5InMessage2.Value != "0" {
		t.Errorf("Expected field5InMessage2 value to be '0', got %v", field5InMessage2.Value)
	}
}
