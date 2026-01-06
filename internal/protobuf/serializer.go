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

	protoContent := s.GenerateProtoSchema(tree)
	protoFile := filepath.Join(tempDir, "message.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		return nil, fmt.Errorf("error writing proto file: %w", err)
	}

	textFormat := s.TreeToTextFormatWithNames(tree)

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

	if node.Type == "message" || len(node.Children) > 0 {
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
			} else if node.Type == "number" {
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
	var builder strings.Builder
	builder.WriteString("syntax = \"proto3\";\n\n")
	builder.WriteString("message Message {\n")

	fieldNum := 1
	s.WriteProtoFields(&builder, tree, &fieldNum)

	builder.WriteString("}\n")
	return builder.String()
}

func (s *Serializer) WriteProtoFields(builder *strings.Builder, node *TreeNode, fieldNum *int) {
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
				s.WriteProtoField(builder, child, fieldNum)
			} else {
				s.WriteProtoField(builder, child, fieldNum)
			}
		}
		return
	}

	s.WriteProtoField(builder, node, fieldNum)
}

func (s *Serializer) WriteProtoField(builder *strings.Builder, node *TreeNode, fieldNum *int) {
	indent := "  "

	if node.Type == "message" || len(node.Children) > 0 {
		messageName := fmt.Sprintf("Message%d", *fieldNum)
		builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, messageName, node.Name, node.FieldNum))

		*fieldNum++
		childFieldNum := 1
		builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
		for _, child := range node.Children {
			s.WriteProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ")
		}
		builder.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		protoType := s.MapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
	}
}

func (s *Serializer) WriteProtoFieldRecursive(builder *strings.Builder, node *TreeNode, fieldNum *int, indent string) {
	if node.Type == "message" || len(node.Children) > 0 {
		messageName := fmt.Sprintf("Message%d", *fieldNum)
		builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, messageName, node.Name, node.FieldNum))

		*fieldNum++
		childFieldNum := 1
		builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
		for _, child := range node.Children {
			s.WriteProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ")
		}
		builder.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		protoType := s.MapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
	}
}

func (s *Serializer) MapTypeToProtoType(ourType string) string {
	switch ourType {
	case "string":
		return "string"
	case "number":
		return "int32"
	case "bool":
		return "bool"
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

	if node.Type == "message" || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%s {\n", node.Name))
		for _, child := range node.Children {
			s.WriteNodeToTextFormatWithNames(builder, child, indent+1)
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
			} else if node.Type == "number" {
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
