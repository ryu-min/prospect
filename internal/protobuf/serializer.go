package protobuf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Serializer struct {
	protocPath string
}

func NewSerializer(protocPath string) *Serializer {
	return &Serializer{
		protocPath: protocPath,
	}
}

func (s *Serializer) SerializeRaw(tree *TreeNode) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "prospect_proto_*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fieldNameMap := make(map[int]string)
	messageCounter := 1
	usedMessageNames := make(map[string]string)
	protoContent := s.GenerateProtoSchemaWithFieldNames(tree, fieldNameMap, &messageCounter, usedMessageNames)
	protoFile := filepath.Join(tempDir, "message.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		return nil, fmt.Errorf("error writing proto file: %w", err)
	}

	textFormat := s.TreeToTextFormatWithFieldNames(tree, fieldNameMap)

	messageName := "Message"
	cmd := exec.Command(s.protocPath, "--encode", messageName, "--proto_path", tempDir, protoFile)
	cmd.Stdin = strings.NewReader(textFormat)

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			return nil, fmt.Errorf("error encoding protobuf: %w\nstderr: %s\ntext format:\n%s\nproto schema:\n%s", err, stderr, textFormat, protoContent)
		}
		return nil, fmt.Errorf("error encoding protobuf: %w", err)
	}

	return output, nil
}

func (s *Serializer) TreeToTextFormat(node *TreeNode) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	s.WriteNodeToTextFormat(&result, node, 0)
	return result.String()
}

func (s *Serializer) WriteNodeToTextFormat(builder *strings.Builder, node *TreeNode, indent int) {
	if node.Name == "root" {
		for _, child := range node.Children {
			s.WriteNodeToTextFormat(builder, child, indent)
		}
		return
	}

	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	if isMessageType(node.Type) || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%d {\n", node.FieldNum))
		for _, child := range node.Children {
			s.WriteNodeToTextFormat(builder, child, indent+1)
		}
		for i := 0; i < indent; i++ {
			builder.WriteString("  ")
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString(fmt.Sprintf("%d: ", node.FieldNum))
		if node.Value != nil {
			if node.Type == "string" {
				builder.WriteString(fmt.Sprintf("\"%s\"", fmt.Sprintf("%v", node.Value)))
			} else if node.Type == "bool" {
				if v, ok := node.Value.(bool); ok {
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				} else {
					valueStr := fmt.Sprintf("%v", node.Value)
					if valueStr == "true" || valueStr == "1" {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				}
			} else if node.Type == "int32" || node.Type == "int64" || node.Type == "uint32" || node.Type == "uint64" || node.Type == "sint32" || node.Type == "sint64" || node.Type == "float" || node.Type == "double" {
				builder.WriteString(fmt.Sprintf("%v", node.Value))
			} else {
				switch v := node.Value.(type) {
				case string:
					if isNumeric(v) {
						builder.WriteString(v)
					} else {
						builder.WriteString(fmt.Sprintf("\"%s\"", v))
					}
				case bool:
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				default:
					builder.WriteString(fmt.Sprintf("%v", v))
				}
			}
		}
		builder.WriteString("\n")
	}
}

func (s *Serializer) GenerateProtoSchema(tree *TreeNode) string {
	fieldNameMap := make(map[int]string)
	messageCounter := 1
	usedMessageNames := make(map[string]string)
	return s.GenerateProtoSchemaWithFieldNames(tree, fieldNameMap, &messageCounter, usedMessageNames)
}

func (s *Serializer) GenerateProtoSchemaWithFieldNames(tree *TreeNode, fieldNameMap map[int]string, messageCounter *int, usedMessageNames map[string]string) string {
	var builder strings.Builder
	builder.WriteString("syntax = \"proto2\";\n\n")
	builder.WriteString("message Message {\n")

	fieldNum := 1
	fieldNameCounter := make(map[string]int)
	s.WriteProtoFields(&builder, tree, &fieldNum, messageCounter, usedMessageNames, fieldNameCounter, fieldNameMap)

	builder.WriteString("}\n")
	return builder.String()
}

func (s *Serializer) WriteProtoFields(builder *strings.Builder, node *TreeNode, fieldNum *int, messageCounter *int, usedMessageNames map[string]string, fieldNameCounter map[string]int, fieldNameMap map[int]string) {
	if node.Name == "root" {
		fieldMap := make(map[int][]*TreeNode)
		for _, child := range node.Children {
			fieldMap[child.FieldNum] = append(fieldMap[child.FieldNum], child)
		}

		processedFields := make(map[int]bool)
		for _, child := range node.Children {
			if processedFields[child.FieldNum] {
				continue
			}
			processedFields[child.FieldNum] = true

			if len(fieldMap[child.FieldNum]) > 1 {
				child.IsRepeated = true
				s.WriteProtoField(builder, child, fieldNum, messageCounter, usedMessageNames, fieldNameCounter, fieldNameMap)
			} else {
				s.WriteProtoField(builder, child, fieldNum, messageCounter, usedMessageNames, fieldNameCounter, fieldNameMap)
			}
		}
		return
	}

	s.WriteProtoField(builder, node, fieldNum, messageCounter, usedMessageNames, fieldNameCounter, fieldNameMap)
}

func (s *Serializer) WriteProtoField(builder *strings.Builder, node *TreeNode, fieldNum *int, messageCounter *int, usedMessageNames map[string]string, fieldNameCounter map[string]int, fieldNameMap map[int]string) {
	indent := "  "

	if isMessageType(node.Type) || len(node.Children) > 0 {
		var messageName string
		alreadyDefined := false
		if existingName, exists := usedMessageNames[node.Name]; exists {
			messageName = existingName
			alreadyDefined = true
		} else {
			messageName = fmt.Sprintf("Message%d", *messageCounter)
			usedMessageNames[node.Name] = messageName
			*messageCounter++
		}

		fieldName := node.Name
		if isMessageType(node.Type) {
			fieldNameCounter[fieldName]++
			if fieldNameCounter[fieldName] > 1 {
				fieldName = fmt.Sprintf("%s_%d", fieldName, fieldNameCounter[fieldName])
			}
		}

		fieldNameMap[node.FieldNum] = fieldName
		builder.WriteString(fmt.Sprintf("%soptional %s %s = %d;\n", indent, messageName, fieldName, node.FieldNum))

		if !alreadyDefined {
			childFieldNum := 1
			builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
			for _, child := range node.Children {
				s.WriteProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ", messageCounter, usedMessageNames)
			}
			builder.WriteString(fmt.Sprintf("%s}\n", indent))
		}
		*fieldNum++
	} else {
		protoType := s.MapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%soptional %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
		*fieldNum++
	}
}

func (s *Serializer) WriteProtoFieldRecursive(builder *strings.Builder, node *TreeNode, fieldNum *int, indent string, messageCounter *int, usedMessageNames map[string]string) {
	if isMessageType(node.Type) || len(node.Children) > 0 {
		var messageName string
		alreadyDefined := false
		if existingName, exists := usedMessageNames[node.Name]; exists {
			messageName = existingName
			alreadyDefined = true
		} else {
			messageName = fmt.Sprintf("Message%d", *messageCounter)
			usedMessageNames[node.Name] = messageName
			*messageCounter++
		}
		builder.WriteString(fmt.Sprintf("%soptional %s %s = %d;\n", indent, messageName, node.Name, node.FieldNum))

		if !alreadyDefined {
			childFieldNum := 1
			builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
			for _, child := range node.Children {
				s.WriteProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ", messageCounter, usedMessageNames)
			}
			builder.WriteString(fmt.Sprintf("%s}\n", indent))
		}
		*fieldNum++
	} else {
		protoType := s.MapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%soptional %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
		*fieldNum++
	}
}

func (s *Serializer) MapTypeToProtoType(ourType string) string {
	switch ourType {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "sint32":
		return "sint32"
	case "sint64":
		return "sint64"
	case "bool":
		return "bool"
	case "float":
		return "float"
	case "double":
		return "double"
	default:
		return "string"
	}
}

func (s *Serializer) TreeToTextFormatWithNames(node *TreeNode) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	s.WriteNodeToTextFormatWithNames(&result, node, 0)
	return result.String()
}

func (s *Serializer) TreeToTextFormatWithFieldNames(node *TreeNode, fieldNameMap map[int]string) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	s.WriteNodeToTextFormatWithFieldNames(&result, node, 0, fieldNameMap)
	return result.String()
}

func (s *Serializer) WriteNodeToTextFormatWithNames(builder *strings.Builder, node *TreeNode, indent int) {
	if node.Name == "root" {
		for _, child := range node.Children {
			s.WriteNodeToTextFormatWithNames(builder, child, indent)
		}
		return
	}

	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	if isMessageType(node.Type) || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%d {\n", node.FieldNum))
		for _, child := range node.Children {
			s.WriteNodeToTextFormatWithNames(builder, child, indent+1)
		}
		for i := 0; i < indent; i++ {
			builder.WriteString("  ")
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString(fmt.Sprintf("%d: ", node.FieldNum))
		if node.Value != nil {
			if node.Type == "string" {
				builder.WriteString(fmt.Sprintf("\"%s\"", fmt.Sprintf("%v", node.Value)))
			} else if node.Type == "bool" {
				if v, ok := node.Value.(bool); ok {
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				} else {
					valueStr := fmt.Sprintf("%v", node.Value)
					if valueStr == "true" || valueStr == "1" {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				}
			} else if node.Type == "int32" || node.Type == "int64" || node.Type == "uint32" || node.Type == "uint64" || node.Type == "sint32" || node.Type == "sint64" || node.Type == "float" || node.Type == "double" {
				builder.WriteString(fmt.Sprintf("%v", node.Value))
			} else {
				switch v := node.Value.(type) {
				case string:
					if isNumeric(v) {
						builder.WriteString(v)
					} else {
						builder.WriteString(fmt.Sprintf("\"%s\"", v))
					}
				case bool:
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				default:
					builder.WriteString(fmt.Sprintf("%v", v))
				}
			}
		}
		builder.WriteString("\n")
	}
}

func (s *Serializer) WriteNodeToTextFormatWithFieldNames(builder *strings.Builder, node *TreeNode, indent int, fieldNameMap map[int]string) {
	if node.Name == "root" {
		for _, child := range node.Children {
			s.WriteNodeToTextFormatWithFieldNames(builder, child, indent, fieldNameMap)
		}
		return
	}

	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	if isMessageType(node.Type) || len(node.Children) > 0 {
		fieldName, exists := fieldNameMap[node.FieldNum]
		if !exists {
			fieldName = node.Name
		}
		builder.WriteString(fmt.Sprintf("%s {\n", fieldName))
		for _, child := range node.Children {
			s.WriteNodeToTextFormatWithFieldNames(builder, child, indent+1, fieldNameMap)
		}
		for i := 0; i < indent; i++ {
			builder.WriteString("  ")
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString(fmt.Sprintf("%s: ", node.Name))
		if node.Value != nil {
			if node.Type == "string" {
				builder.WriteString(fmt.Sprintf("\"%s\"", fmt.Sprintf("%v", node.Value)))
			} else if node.Type == "bool" {
				if v, ok := node.Value.(bool); ok {
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				} else {
					valueStr := fmt.Sprintf("%v", node.Value)
					if valueStr == "true" || valueStr == "1" {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				}
			} else if node.Type == "int32" || node.Type == "int64" || node.Type == "uint32" || node.Type == "uint64" || node.Type == "sint32" || node.Type == "sint64" || node.Type == "float" || node.Type == "double" {
				builder.WriteString(fmt.Sprintf("%v", node.Value))
			} else {
				switch v := node.Value.(type) {
				case string:
					if isNumeric(v) {
						builder.WriteString(v)
					} else {
						builder.WriteString(fmt.Sprintf("\"%s\"", v))
					}
				case bool:
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				default:
					builder.WriteString(fmt.Sprintf("%v", v))
				}
			}
		}
		builder.WriteString("\n")
	}
}
