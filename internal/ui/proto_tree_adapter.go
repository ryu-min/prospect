package ui

import (
	"fmt"
	"strconv"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var dontAskTypeChangeConfirmation bool
var dontAskFieldTypeSyncConfirmation bool

type protoTreeAdapter struct {
	tree        *protobuf.TreeNode
	editWidgets map[widget.TreeNodeID]*protoFieldEditor
	window      fyne.Window
	treeWidget  *widget.Tree
}

func newProtoTreeAdapter(tree *protobuf.TreeNode) *protoTreeAdapter {
	return &protoTreeAdapter{
		tree:        tree,
		editWidgets: make(map[widget.TreeNodeID]*protoFieldEditor),
		window:      nil,
	}
}

func (a *protoTreeAdapter) SetWindow(window fyne.Window) {
	a.window = window
}

func (a *protoTreeAdapter) SetTreeWidget(treeWidget *widget.Tree) {
	a.treeWidget = treeWidget
}

func (a *protoTreeAdapter) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		return nil
	}

	children := make([]widget.TreeNodeID, 0, len(node.Children))
	for i := range node.Children {
		if actualUID == "root" {
			childUID := fmt.Sprintf("%d", i)
			children = append(children, childUID)
		} else {
			childUID := fmt.Sprintf("%s:%d", uid, i)
			children = append(children, childUID)
		}
	}
	return children
}

func (a *protoTreeAdapter) IsBranch(uid widget.TreeNodeID) bool {
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		return false
	}
	return a.isMessageType(node.Type) || len(node.Children) > 0
}

func (a *protoTreeAdapter) CreateNode(branch bool) fyne.CanvasObject {
	messageTypes := a.getAllMessageTypes()
	ew := newProtoFieldEditor("", a, messageTypes)
	return ew
}

func (a *protoTreeAdapter) UpdateNode(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
	actualUID := uid
	if uid == "" {
		actualUID = "root"
	}

	node := a.getNodeByUID(actualUID)
	if node == nil {
		if label, ok := obj.(*widget.Label); ok {
			label.SetText(fmt.Sprintf("ERROR: Node not found ('%s')", uid))
		}
		return
	}

	if node.Name == "root" {
		if label, ok := obj.(*widget.Label); ok {
			text := "Proto Root"
			if len(node.Children) == 0 {
				text = "Proto Root (no data)"
			}
			label.SetText(text)
		}
		return
	}

	if editWidget, ok := obj.(*protoFieldEditor); ok {
		editWidget.uid = actualUID
		a.editWidgets[actualUID] = editWidget

		var nameText string
		if a.isMessageType(node.Type) {
			nameText = node.Name
		} else {
			nameText = fmt.Sprintf("field_%d", node.FieldNum)
		}
		editWidget.nameLabel.SetText(nameText)

		allTypes := a.getAvailableTypesForNode(node)

		editWidget.availableTypes = allTypes
		editWidget.typeCombo.Options = allTypes

		typeText := node.Type

		editWidget.typeCombo.OnChanged = nil
		editWidget.typeCombo.SetSelected(typeText)

		editWidget.typeCombo.OnChanged = func(selectedType string) {
			if selectedType == "" || selectedType == node.Type {
				return
			}
			a.handleTypeChange(actualUID, node.Type, selectedType)
		}

		if a.isMessageType(node.Type) {
			editWidget.SetEntryVisible(false)
			editWidget.entry.SetText("")
			editWidget.entry.OnChanged = nil
		} else {
			editWidget.SetEntryVisible(true)
			editWidget.entry.Enable()
			valueStr := a.nodeValueToString(node)

			editWidget.entry.OnChanged = nil
			editWidget.entry.SetText(valueStr)

			lastValidValue := valueStr
			var isUpdating bool

			fieldType := node.Type
			editWidget.entry.OnChanged = func(value string) {
				if isUpdating {
					return
				}

				if !a.validateValue(value, fieldType) {
					isUpdating = true
					editWidget.entry.SetText(lastValidValue)
					isUpdating = false
					return
				}

				lastValidValue = value
				a.updateNodeValue(actualUID, value, fieldType)
			}
		}
		editWidget.Refresh()
		return
	}
}

func (a *protoTreeAdapter) updateEntryValidation(uid widget.TreeNodeID, newType string) {
	editWidget, ok := a.editWidgets[uid]
	if !ok {
		return
	}

	if a.isMessageType(newType) {
		editWidget.SetEntryVisible(false)
		editWidget.entry.SetText("")
		editWidget.entry.OnChanged = nil
		return
	}

	editWidget.SetEntryVisible(true)
	editWidget.entry.Enable()

	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	currentValue := editWidget.entry.Text
	lastValidValue := currentValue
	var isUpdating bool

	editWidget.entry.OnChanged = nil
	editWidget.entry.OnChanged = func(value string) {
		if isUpdating {
			return
		}

		if !a.validateValue(value, newType) {
			isUpdating = true
			editWidget.entry.SetText(lastValidValue)
			isUpdating = false
			return
		}

		lastValidValue = value
		a.updateNodeValue(uid, value, newType)
	}
}

func (a *protoTreeAdapter) nodeValueToString(node *protobuf.TreeNode) string {
	if node.Value == nil {
		return ""
	}

	switch v := node.Value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (a *protoTreeAdapter) validateValue(value string, fieldType string) bool {
	if value == "" {
		return true
	}

	if a.isMessageType(fieldType) {
		return true
	}

	switch fieldType {
	case "string":
		return true
	case "int32", "int64", "sint32", "sint64":
		if value == "-" {
			return true
		}
		trimmed := strings.TrimSpace(value)
		_, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			if trimmed == "-" {
				return true
			}
			return false
		}
		return true
	case "uint32", "uint64":
		trimmed := strings.TrimSpace(value)
		_, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil {
			return false
		}
		return true
	case "float", "double":
		if value == "" {
			return true
		}
		if value == "-" {
			return true
		}
		if value == "." {
			return true
		}
		if value == "-." {
			return true
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "-" || trimmed == "." || trimmed == "-." {
			return true
		}
		if strings.HasSuffix(trimmed, ".") {
			trimmed = strings.TrimSuffix(trimmed, ".")
			if trimmed == "" || trimmed == "-" {
				return true
			}
		}
		_, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return false
		}
		return true
	case "bool":
		trimmed := strings.TrimSpace(value)
		return trimmed == "0" || trimmed == "1" || trimmed == ""
	default:
		return true
	}
}

func (a *protoTreeAdapter) detectTypeChange(oldType string, valueStr string) (newType string, changed bool) {
	if valueStr == "" {
		return oldType, false
	}

	trimmed := strings.TrimSpace(valueStr)

	if oldType == "bool" {
		if trimmed == "0" || trimmed == "1" {
			return "bool", false
		}
	}

	if oldType == "int32" || oldType == "int64" || oldType == "sint32" || oldType == "sint64" {
		if trimmed == "-" {
			return oldType, false
		}
		_, err := strconv.ParseInt(trimmed, 10, 64)
		if err == nil {
			return oldType, false
		}
		if trimmed == "-" || trimmed == "." || trimmed == "-." {
			return oldType, false
		}
		trimmedForFloat := trimmed
		if strings.HasSuffix(trimmedForFloat, ".") {
			trimmedForFloat = strings.TrimSuffix(trimmedForFloat, ".")
		}
		_, errFloat := strconv.ParseFloat(trimmedForFloat, 64)
		if errFloat == nil {
			return "float", true
		}
		return "string", true
	}

	if oldType == "uint32" || oldType == "uint64" {
		_, err := strconv.ParseUint(trimmed, 10, 64)
		if err == nil {
			return oldType, false
		}
		if trimmed == "." || trimmed == "-." {
			return oldType, false
		}
		trimmedForFloat := trimmed
		if strings.HasSuffix(trimmedForFloat, ".") {
			trimmedForFloat = strings.TrimSuffix(trimmedForFloat, ".")
		}
		_, errFloat := strconv.ParseFloat(trimmedForFloat, 64)
		if errFloat == nil {
			return "float", true
		}
		return "string", true
	}

	if oldType == "float" || oldType == "double" {
		if trimmed == "-" || trimmed == "." || trimmed == "-." {
			return oldType, false
		}
		trimmedForFloat := trimmed
		if strings.HasSuffix(trimmedForFloat, ".") {
			trimmedForFloat = strings.TrimSuffix(trimmedForFloat, ".")
		}
		_, errFloat := strconv.ParseFloat(trimmedForFloat, 64)
		if errFloat == nil {
			return oldType, false
		}
		_, errInt := strconv.ParseInt(trimmed, 10, 64)
		if errInt == nil {
			return "int32", true
		}
		return "string", true
	}

	if oldType == "string" {
		return oldType, false
	}

	if trimmed == "0" || trimmed == "1" {
		return "bool", oldType != "bool"
	}
	if trimmed == "-" || trimmed == "." || trimmed == "-." {
		return oldType, false
	}
	trimmedForNumber := trimmed
	if strings.HasSuffix(trimmedForNumber, ".") {
		trimmedForNumber = strings.TrimSuffix(trimmedForNumber, ".")
	}
	_, errInt := strconv.ParseInt(trimmedForNumber, 10, 64)
	_, errFloat := strconv.ParseFloat(trimmedForNumber, 64)
	if errInt == nil {
		return "int32", oldType != "int32" && oldType != "int64" && oldType != "sint32" && oldType != "sint64" && oldType != "uint32" && oldType != "uint64"
	}
	if errFloat == nil {
		return "float", oldType != "float" && oldType != "double"
	}
	return "string", oldType != "string"
}

func (a *protoTreeAdapter) showTypeChangeDialog(oldType, newType string, onConfirm func(), onCancel func()) {
	if dontAskTypeChangeConfirmation {
		onConfirm()
		return
	}

	if a.window == nil {
		onCancel()
		return
	}

	message := fmt.Sprintf("Changing field type from '%s' to '%s' will clear the field value.\n\nDo you want to continue?", oldType, newType)
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	checkbox := widget.NewCheck("Don't ask again", func(checked bool) {
		dontAskTypeChangeConfirmation = checked
	})

	content := container.NewVBox(
		label,
		checkbox,
	)

	confirmDialog := dialog.NewCustomConfirm("Type Change Confirmation", "Yes", "No", content, func(confirmed bool) {
		if confirmed {
			onConfirm()
		} else {
			onCancel()
		}
	}, a.window)
	confirmDialog.Show()
}

func (a *protoTreeAdapter) handleTypeChange(uid widget.TreeNodeID, oldType, newType string) {
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	isMessageType := a.isMessageType(newType)
	isOldMessageType := a.isMessageType(oldType)

	if isMessageType && !isOldMessageType {
		parentMessage := a.findParentMessage(node)

		if parentMessage != nil {
			affectedFields := a.findFieldsWithSameFieldNumInMessageType(node, parentMessage.Type, node.FieldNum)
			if len(affectedFields) > 0 {
				a.showFieldTypeSyncDialog(
					node.FieldNum,
					oldType,
					newType,
					len(affectedFields),
					func() {
						sourceMessage := a.findMessageByName(newType)
						var finalMessageType string

						if sourceMessage != nil {
							finalMessageType = newType
						} else {
							messageCounter := a.countMessages(a.tree)
							finalMessageType = fmt.Sprintf("message_%d", messageCounter+1)
						}

						if len(node.Children) > 0 {
							node.Type = finalMessageType
							if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
								node.Name = fmt.Sprintf("field_%d", node.FieldNum)
							}
							node.Value = nil
						} else {
							if sourceMessage != nil {
								node.Type = finalMessageType
								if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
									node.Name = fmt.Sprintf("field_%d", node.FieldNum)
								}
								node.Value = nil
								node.Children = a.copyMessageChildren(sourceMessage)
							} else {
								node.Type = finalMessageType
								if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
									node.Name = fmt.Sprintf("field_%d", node.FieldNum)
								}
								node.Value = nil
								node.Children = make([]*protobuf.TreeNode, 0)
							}
						}

						for _, field := range affectedFields {
							if len(field.Children) > 0 {
								field.Type = finalMessageType
								if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
									field.Name = fmt.Sprintf("field_%d", field.FieldNum)
								}
								field.Value = nil
							} else {
								if sourceMessage != nil {
									field.Type = finalMessageType
									if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
										field.Name = fmt.Sprintf("field_%d", field.FieldNum)
									}
									field.Value = nil
									field.Children = a.copyMessageChildren(sourceMessage)
								} else {
									field.Type = finalMessageType
									if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
										field.Name = fmt.Sprintf("field_%d", field.FieldNum)
									}
									field.Value = nil
									field.Children = make([]*protobuf.TreeNode, 0)
								}
							}
						}

						delete(a.editWidgets, uid)
						if a.treeWidget != nil {
							a.treeWidget.Refresh()
						}
					},
					func() {
						if editWidget, ok := a.editWidgets[uid]; ok {
							editWidget.typeCombo.SetSelected(oldType)
						}
					},
				)
				return
			}
		}

		if len(node.Children) > 0 {
			node.Type = newType
			if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
				node.Name = fmt.Sprintf("field_%d", node.FieldNum)
			}
			node.Value = nil
		} else {
			sourceMessage := a.findMessageByName(newType)
			if sourceMessage != nil {
				node.Type = newType
				if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
					node.Name = fmt.Sprintf("field_%d", node.FieldNum)
				}
				node.Value = nil
				node.Children = a.copyMessageChildren(sourceMessage)
			} else {
				messageCounter := a.countMessages(a.tree)
				messageName := fmt.Sprintf("message_%d", messageCounter+1)
				node.Type = messageName
				if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
					node.Name = fmt.Sprintf("field_%d", node.FieldNum)
				}
				node.Value = nil
				node.Children = make([]*protobuf.TreeNode, 0)
			}
		}
		delete(a.editWidgets, uid)
		if a.treeWidget != nil {
			a.treeWidget.Refresh()
		}
		return
	}

	if isMessageType && isOldMessageType {
		parentMessage := a.findParentMessage(node)

		if parentMessage != nil {
			affectedFields := a.findFieldsWithSameFieldNumInMessageType(node, parentMessage.Type, node.FieldNum)
			if len(affectedFields) > 0 {
				a.showFieldTypeSyncDialog(
					node.FieldNum,
					oldType,
					newType,
					len(affectedFields),
					func() {
						sourceMessage := a.findMessageByName(newType)
						var finalMessageType string

						if sourceMessage != nil {
							finalMessageType = newType
						} else {
							messageCounter := a.countMessages(a.tree)
							finalMessageType = fmt.Sprintf("message_%d", messageCounter+1)
						}

						if len(node.Children) > 0 {
							node.Type = finalMessageType
							if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
								node.Name = fmt.Sprintf("field_%d", node.FieldNum)
							}
							node.Value = nil
						} else {
							if sourceMessage != nil {
								node.Type = finalMessageType
								if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
									node.Name = fmt.Sprintf("field_%d", node.FieldNum)
								}
								node.Value = nil
								node.Children = a.copyMessageChildren(sourceMessage)
							} else {
								node.Type = finalMessageType
								if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
									node.Name = fmt.Sprintf("field_%d", node.FieldNum)
								}
								node.Value = nil
								node.Children = make([]*protobuf.TreeNode, 0)
							}
						}

						for _, field := range affectedFields {
							if len(field.Children) > 0 {
								field.Type = finalMessageType
								if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
									field.Name = fmt.Sprintf("field_%d", field.FieldNum)
								}
								field.Value = nil
							} else {
								if sourceMessage != nil {
									field.Type = finalMessageType
									if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
										field.Name = fmt.Sprintf("field_%d", field.FieldNum)
									}
									field.Value = nil
									field.Children = a.copyMessageChildren(sourceMessage)
								} else {
									field.Type = finalMessageType
									if field.Name == "" || !strings.HasPrefix(field.Name, "field_") {
										field.Name = fmt.Sprintf("field_%d", field.FieldNum)
									}
									field.Value = nil
									field.Children = make([]*protobuf.TreeNode, 0)
								}
							}
						}

						delete(a.editWidgets, uid)
						if a.treeWidget != nil {
							a.treeWidget.Refresh()
						}
					},
					func() {
						if editWidget, ok := a.editWidgets[uid]; ok {
							editWidget.typeCombo.SetSelected(oldType)
						}
					},
				)
				return
			}
		}

		if len(node.Children) > 0 {
			node.Type = newType
			if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
				node.Name = fmt.Sprintf("field_%d", node.FieldNum)
			}
			node.Value = nil
		} else {
			sourceMessage := a.findMessageByName(newType)
			if sourceMessage != nil {
				node.Type = newType
				if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
					node.Name = fmt.Sprintf("field_%d", node.FieldNum)
				}
				node.Value = nil
				node.Children = a.copyMessageChildren(sourceMessage)
			} else {
				messageCounter := a.countMessages(a.tree)
				messageName := fmt.Sprintf("message_%d", messageCounter+1)
				node.Type = messageName
				if node.Name == "" || !strings.HasPrefix(node.Name, "field_") {
					node.Name = fmt.Sprintf("field_%d", node.FieldNum)
				}
				node.Value = nil
				node.Children = make([]*protobuf.TreeNode, 0)
			}
		}
		delete(a.editWidgets, uid)
		if a.treeWidget != nil {
			a.treeWidget.Refresh()
		}
		return
	}

	canSeamlessChange := false
	valueStr := a.nodeValueToString(node)
	isIntegerType := func(t string) bool {
		return t == "int32" || t == "int64" || t == "uint32" || t == "uint64" || t == "sint32" || t == "sint64"
	}
	isFloatType := func(t string) bool {
		return t == "float" || t == "double"
	}

	if (oldType == "bool" && (isIntegerType(newType) || isFloatType(newType))) ||
		((isIntegerType(oldType) || isFloatType(oldType)) && newType == "bool") {
		if valueStr == "0" || valueStr == "1" {
			canSeamlessChange = true
		}
	}

	if canSeamlessChange {
		parentMessage := a.findParentMessage(node)
		var affectedFields []*protobuf.TreeNode
		if parentMessage != nil {
			affectedFields = a.findFieldsWithSameFieldNumInMessageType(node, parentMessage.Type, node.FieldNum)
		}

		node.Type = newType
		if oldType == "bool" && (isIntegerType(newType) || isFloatType(newType)) {
			node.Value = valueStr
		} else if (isIntegerType(oldType) || isFloatType(oldType)) && newType == "bool" {
			if valueStr == "1" {
				node.Value = true
			} else if valueStr == "0" {
				node.Value = false
			}
		}

		for _, field := range affectedFields {
			field.Type = newType
			if oldType == "bool" && (isIntegerType(newType) || isFloatType(newType)) {
				fieldValueStr := a.nodeValueToString(field)
				field.Value = fieldValueStr
			} else if (isIntegerType(oldType) || isFloatType(oldType)) && newType == "bool" {
				fieldValueStr := a.nodeValueToString(field)
				if fieldValueStr == "1" {
					field.Value = true
				} else if fieldValueStr == "0" {
					field.Value = false
				}
			}
		}

		if editWidget, ok := a.editWidgets[uid]; ok {
			editWidget.typeCombo.SetSelected(newType)
			newValueStr := a.nodeValueToString(node)
			editWidget.entry.SetText(newValueStr)
			a.updateEntryValidation(uid, newType)
		}

		if len(affectedFields) > 0 && a.treeWidget != nil {
			a.treeWidget.Refresh()
		}
		return
	}

	oldValue := node.Value
	parentMessage := a.findParentMessage(node)

	if parentMessage != nil {
		affectedFields := a.findFieldsWithSameFieldNumInMessageType(node, parentMessage.Type, node.FieldNum)
		if len(affectedFields) > 0 {
			a.showFieldTypeSyncDialog(
				node.FieldNum,
				oldType,
				newType,
				len(affectedFields),
				func() {
					node.Value = nil
					node.Type = newType
					if editWidget, ok := a.editWidgets[uid]; ok {
						editWidget.entry.SetText("")
						editWidget.typeCombo.SetSelected(newType)
						a.updateEntryValidation(uid, newType)
					}

					for _, field := range affectedFields {
						field.Value = nil
						field.Type = newType
					}

					if a.treeWidget != nil {
						a.treeWidget.Refresh()
					}
				},
				func() {
					node.Type = oldType
					node.Value = oldValue
					if editWidget, ok := a.editWidgets[uid]; ok {
						valueStr := a.nodeValueToString(node)
						editWidget.entry.SetText(valueStr)
						editWidget.typeCombo.SetSelected(oldType)
						a.updateEntryValidation(uid, oldType)
					}
				},
			)
			return
		}
	}

	a.showTypeChangeDialog(
		oldType,
		newType,
		func() {
			node.Value = nil
			node.Type = newType
			if editWidget, ok := a.editWidgets[uid]; ok {
				editWidget.entry.SetText("")
				editWidget.typeCombo.SetSelected(newType)
				a.updateEntryValidation(uid, newType)
			}
		},
		func() {
			node.Type = oldType
			node.Value = oldValue
			if editWidget, ok := a.editWidgets[uid]; ok {
				valueStr := a.nodeValueToString(node)
				editWidget.entry.SetText(valueStr)
				editWidget.typeCombo.SetSelected(oldType)
				a.updateEntryValidation(uid, oldType)
			}
		},
	)
}

func (a *protoTreeAdapter) isMessageType(typeName string) bool {
	if typeName == "message" {
		return true
	}
	if strings.HasPrefix(typeName, "message_") {
		return true
	}
	return false
}

func (a *protoTreeAdapter) updateNodeValue(uid widget.TreeNodeID, valueStr string, fieldType string) {
	node := a.getNodeByUID(uid)
	if node == nil {
		return
	}

	newType, typeChanged := a.detectTypeChange(fieldType, valueStr)

	if typeChanged {
		oldValue := node.Value
		oldType := node.Type

		a.showTypeChangeDialog(
			oldType,
			newType,
			func() {
				node.Value = nil
				node.Type = newType
				if editWidget, ok := a.editWidgets[uid]; ok {
					editWidget.entry.SetText("")
					editWidget.typeCombo.SetSelected(newType)
					a.updateEntryValidation(uid, newType)
				}
			},
			func() {
				node.Value = oldValue
				node.Type = oldType
				if editWidget, ok := a.editWidgets[uid]; ok {
					valueStr := a.nodeValueToString(node)
					editWidget.entry.SetText(valueStr)
					editWidget.typeCombo.SetSelected(oldType)
				}
			},
		)
		return
	}

	if newType != fieldType {
		node.Type = newType
		if editWidget, ok := a.editWidgets[uid]; ok {
			editWidget.typeCombo.SetSelected(newType)
		}
	}

	switch newType {
	case "string":
		node.Value = valueStr
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "float", "double":
		node.Value = valueStr
	case "bool":
		trimmed := strings.TrimSpace(valueStr)
		if trimmed == "1" {
			node.Value = true
		} else if trimmed == "0" {
			node.Value = false
		} else {
			node.Value = valueStr
		}
	default:
		node.Value = valueStr
	}
}

func (a *protoTreeAdapter) getNodeByUID(uid widget.TreeNodeID) *protobuf.TreeNode {
	if a.tree == nil {
		return nil
	}

	if uid == "" || uid == "root" {
		return a.tree
	}

	var parts []string
	if strings.HasPrefix(uid, "root:") {
		parts = splitUID(uid)
		if len(parts) == 0 || parts[0] != "root" {
			return nil
		}
	} else {
		parts = splitUID(uid)
		if len(parts) == 0 {
			return nil
		}
		parts = append([]string{"root"}, parts...)
	}

	current := a.tree
	for i := 1; i < len(parts); i++ {
		idx := parseInt(parts[i])
		if idx < 0 || idx >= len(current.Children) {
			return nil
		}
		current = current.Children[idx]
	}

	return current
}

func (a *protoTreeAdapter) getAvailableTypesForNode(node *protobuf.TreeNode) []string {
	messageTypes := a.getAllMessageTypes()
	baseTypes := []string{"string", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "bool", "float", "double"}
	allTypes := make([]string, 0, len(baseTypes)+len(messageTypes))
	allTypes = append(allTypes, baseTypes...)

	parentMessage := a.findParentMessage(node)
	var excludedMessageType string
	if parentMessage != nil {
		excludedMessageType = parentMessage.Type
	}

	for _, msgType := range messageTypes {
		if msgType != excludedMessageType {
			allTypes = append(allTypes, msgType)
		}
	}

	typeExists := false
	for _, t := range allTypes {
		if t == node.Type {
			typeExists = true
			break
		}
	}
	if !typeExists && node.Type != "" {
		allTypes = append(allTypes, node.Type)
	}

	return allTypes
}

func (a *protoTreeAdapter) getAllMessageTypes() []string {
	if a.tree == nil {
		return []string{}
	}

	messageTypes := make(map[string]bool)
	a.collectMessageTypes(a.tree, messageTypes)

	result := make([]string, 0, len(messageTypes))
	for msgType := range messageTypes {
		if msgType != "root" && msgType != "" {
			result = append(result, msgType)
		}
	}

	return result
}

func (a *protoTreeAdapter) collectMessageTypes(node *protobuf.TreeNode, messageTypes map[string]bool) {
	if node == nil {
		return
	}

	if a.isMessageType(node.Type) && len(node.Children) > 0 && node.Name != "root" {
		messageTypes[node.Type] = true
	}

	for _, child := range node.Children {
		a.collectMessageTypes(child, messageTypes)
	}
}

func (a *protoTreeAdapter) countMessages(node *protobuf.TreeNode) int {
	if node == nil {
		return 0
	}

	count := 0
	if a.isMessageType(node.Type) && node.Name != "root" {
		count = 1
	}

	for _, child := range node.Children {
		count += a.countMessages(child)
	}

	return count
}

func (a *protoTreeAdapter) findMessageByName(name string) *protobuf.TreeNode {
	if a.tree == nil {
		return nil
	}
	return a.findMessageByNameRecursive(a.tree, name)
}

func (a *protoTreeAdapter) findMessageByNameRecursive(node *protobuf.TreeNode, name string) *protobuf.TreeNode {
	if node == nil {
		return nil
	}

	if a.isMessageType(node.Type) && node.Type == name {
		return node
	}

	for _, child := range node.Children {
		if found := a.findMessageByNameRecursive(child, name); found != nil {
			return found
		}
	}

	return nil
}

func (a *protoTreeAdapter) copyMessageChildren(source *protobuf.TreeNode) []*protobuf.TreeNode {
	if source == nil {
		return make([]*protobuf.TreeNode, 0)
	}

	children := make([]*protobuf.TreeNode, 0, len(source.Children))
	for _, child := range source.Children {
		copied := &protobuf.TreeNode{
			Name:       child.Name,
			Type:       child.Type,
			Value:      child.Value,
			FieldNum:   child.FieldNum,
			IsRepeated: child.IsRepeated,
			Children:   a.copyMessageChildren(child),
		}
		children = append(children, copied)
	}

	return children
}

func splitUID(uid widget.TreeNodeID) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range uid {
		if char == ':' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func parseInt(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

func (a *protoTreeAdapter) findParentMessage(node *protobuf.TreeNode) *protobuf.TreeNode {
	if a.tree == nil || node == nil {
		return nil
	}

	var findParent func(*protobuf.TreeNode, *protobuf.TreeNode) *protobuf.TreeNode
	findParent = func(current *protobuf.TreeNode, target *protobuf.TreeNode) *protobuf.TreeNode {
		if current == nil {
			return nil
		}

		if current == target {
			return nil
		}

		for _, child := range current.Children {
			if child == target {
				if a.isMessageType(current.Type) {
					return current
				}
				parentMsg := findParent(a.tree, current)
				if parentMsg != nil {
					return parentMsg
				}
				return nil
			}
			if found := findParent(child, target); found != nil {
				return found
			}
		}
		return nil
	}

	return findParent(a.tree, node)
}

func (a *protoTreeAdapter) findFieldsWithSameFieldNumInMessageType(node *protobuf.TreeNode, messageType string, fieldNum int) []*protobuf.TreeNode {
	if a.tree == nil || node == nil {
		return []*protobuf.TreeNode{}
	}

	var result []*protobuf.TreeNode

	var findFields func(*protobuf.TreeNode)
	findFields = func(current *protobuf.TreeNode) {
		if current == nil {
			return
		}

		if a.isMessageType(current.Type) && current.Type == messageType {
			for _, child := range current.Children {
				if child.FieldNum == fieldNum && child != node {
					result = append(result, child)
				}
			}
		}

		for _, child := range current.Children {
			findFields(child)
		}
	}

	findFields(a.tree)
	return result
}

func (a *protoTreeAdapter) showFieldTypeSyncDialog(fieldNum int, oldType, newType string, affectedCount int, onConfirm func(), onCancel func()) {
	if dontAskFieldTypeSyncConfirmation {
		onConfirm()
		return
	}

	if a.window == nil {
		onCancel()
		return
	}

	message := fmt.Sprintf("Changing field type from '%s' to '%s' for field_%d will also change the type in %d other field(s) with the same field number in messages of the same type.\n\nDo you want to continue?", oldType, newType, fieldNum, affectedCount)
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	checkbox := widget.NewCheck("Don't ask again", func(checked bool) {
		dontAskFieldTypeSyncConfirmation = checked
	})

	content := container.NewVBox(
		label,
		checkbox,
	)

	confirmDialog := dialog.NewCustomConfirm("Field Type Synchronization", "Yes", "No", content, func(confirmed bool) {
		if confirmed {
			onConfirm()
		} else {
			onCancel()
		}
	}, a.window)
	confirmDialog.Show()
}

func createProtoTree(tree *protobuf.TreeNode) *widget.Tree {
	if tree == nil {
		adapter := newProtoTreeAdapter(&protobuf.TreeNode{
			Name:     "root",
			Type:     "message",
			Children: make([]*protobuf.TreeNode, 0),
		})
		treeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
		adapter.SetTreeWidget(treeWidget)
		return treeWidget
	}

	adapter := newProtoTreeAdapter(tree)
	treeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
	adapter.SetTreeWidget(treeWidget)

	treeWidget.OpenBranch("root")

	return treeWidget
}
