package protobuf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Parser struct {
	protocPath string
}

func NewParser() (*Parser, error) {
	protocPath, err := findProtoc()
	if err != nil {
		return nil, fmt.Errorf("protoc не найден: %w", err)
	}

	return &Parser{
		protocPath: protocPath,
	}, nil
}

func findProtoc() (string, error) {
	cmd := exec.Command("protoc", "--version")
	if err := cmd.Run(); err == nil {
		return "protoc", nil
	}

	possiblePaths := []string{
		"C:\\protoc\\bin\\protoc.exe",
		"C:\\Program Files\\protoc\\bin\\protoc.exe",
		os.Getenv("LOCALAPPDATA") + "\\scoop\\apps\\protobuf\\current\\bin\\protoc.exe",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("protoc не найден в системе")
}

func CheckProtoc() error {
	parser, err := NewParser()
	if err != nil {
		return err
	}

	cmd := exec.Command(parser.protocPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("protoc не работает: %w", err)
	}

	log.Printf("protoc найден: %s", strings.TrimSpace(string(output)))
	return nil
}

func (p *Parser) ParseRaw(data []byte) (*TreeNode, error) {
	cmd := exec.Command(p.protocPath, "--decode_raw")
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования protobuf: %w", err)
	}

	outputStr := string(output)
	return p.parseProtocOutput(outputStr)
}

func (p *Parser) parseProtocOutput(output string) (*TreeNode, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || strings.TrimSpace(output) == "" {
		return &TreeNode{
			Name:     "root",
			Type:     "message",
			Children: make([]*TreeNode, 0),
		}, nil
	}

	root := &TreeNode{
		Name:     "root",
		Type:     "message",
		Children: make([]*TreeNode, 0),
	}

	stack := []*TreeNode{root}
	stackIndents := []int{-1}

	fieldCounts := make(map[string]int)

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		indent := getIndentLevel(originalLine)
		trimmedLine := line

		if trimmedLine == "}" {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
				stackIndents = stackIndents[:len(stackIndents)-1]
			}
			continue
		}

		for len(stack) > 1 && indent <= stackIndents[len(stackIndents)-1] {
			stack = stack[:len(stack)-1]
			stackIndents = stackIndents[:len(stackIndents)-1]
		}

		if strings.HasSuffix(trimmedLine, "{") {
			fieldPart := strings.TrimSpace(strings.TrimSuffix(trimmedLine, "{"))
			fieldNum := parseInt(fieldPart)
			node := &TreeNode{
				Name:       fmt.Sprintf("field_%d", fieldNum),
				Type:       "message",
				FieldNum:   fieldNum,
				Children:   make([]*TreeNode, 0),
				IsRepeated: false,
			}

			if len(stack) > 0 {
				stack[len(stack)-1].AddChild(node)
			}
			stack = append(stack, node)
			stackIndents = append(stackIndents, indent)
			continue
		}

		node := p.parseLine(trimmedLine)
		if node != nil {
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				fieldKey := fmt.Sprintf("%p_%d", parent, node.FieldNum)
				fieldCounts[fieldKey]++
				if fieldCounts[fieldKey] > 1 {
					node.IsRepeated = true
					for _, child := range parent.Children {
						if child.FieldNum == node.FieldNum {
							child.IsRepeated = true
						}
					}
				}
				parent.AddChild(node)
			}
		}
	}

	return root, nil
}

func getIndentLevel(line string) int {
	indent := 0
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indent++
		} else {
			break
		}
	}
	return indent
}

func (p *Parser) parseLine(line string) *TreeNode {
	line = strings.TrimSpace(line)
	if line == "" || line == "{" || line == "}" || strings.HasSuffix(line, "{") {
		return nil
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	fieldNum := parseInt(strings.TrimSpace(parts[0]))
	valueStr := strings.TrimSpace(parts[1])

	node := &TreeNode{
		FieldNum: fieldNum,
		Name:     fmt.Sprintf("field_%d", fieldNum),
	}

	if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
		node.Type = "string"
		node.Value = strings.Trim(valueStr, "\"")
	} else if isNumeric(valueStr) {
		node.Type = "number"
		node.Value = valueStr
		if valueStr == "0" || valueStr == "1" {
			node.Type = "bool"
			if valueStr == "1" {
				node.Value = true
			} else {
				node.Value = false
			}
		}
	} else {
		node.Type = "unknown"
		node.Value = valueStr
	}

	return node
}

func parseInt(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

func isNumeric(s string) bool {
	_, err := fmt.Sscanf(s, "%d", new(int))
	return err == nil
}

func (p *Parser) ApplySchema(tree *TreeNode, schemaPath string) (*TreeNode, error) {
	return tree, nil
}

func (n *TreeNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}

func (p *Parser) SerializeRaw(tree *TreeNode) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "prospect_proto_*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	protoContent := p.generateProtoSchema(tree)
	protoFile := filepath.Join(tempDir, "message.proto")
	if err := os.WriteFile(protoFile, []byte(protoContent), 0644); err != nil {
		return nil, fmt.Errorf("error writing proto file: %w", err)
	}

	textFormat := p.treeToTextFormatWithNames(tree)

	messageName := "Message"
	cmd := exec.Command(p.protocPath, "--encode", messageName, "--proto_path", tempDir, protoFile)
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

func (p *Parser) treeToTextFormat(node *TreeNode) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	p.writeNodeToTextFormat(&result, node, 0)
	return result.String()
}

func (p *Parser) writeNodeToTextFormat(builder *strings.Builder, node *TreeNode, indent int) {
	if node.Name == "root" {
		for _, child := range node.Children {
			p.writeNodeToTextFormat(builder, child, indent)
		}
		return
	}

	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	if node.Type == "message" || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%d {\n", node.FieldNum))
		for _, child := range node.Children {
			p.writeNodeToTextFormat(builder, child, indent+1)
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

func (p *Parser) generateProtoSchema(tree *TreeNode) string {
	var builder strings.Builder
	builder.WriteString("syntax = \"proto3\";\n\n")
	builder.WriteString("message Message {\n")

	fieldNum := 1
	p.writeProtoFields(&builder, tree, &fieldNum)

	builder.WriteString("}\n")
	return builder.String()
}

func (p *Parser) writeProtoFields(builder *strings.Builder, node *TreeNode, fieldNum *int) {
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
				p.writeProtoField(builder, child, fieldNum)
			} else {
				p.writeProtoField(builder, child, fieldNum)
			}
		}
		return
	}

	p.writeProtoField(builder, node, fieldNum)
}

func (p *Parser) writeProtoField(builder *strings.Builder, node *TreeNode, fieldNum *int) {
	indent := "  "

	if node.Type == "message" || len(node.Children) > 0 {
		messageName := fmt.Sprintf("Message%d", *fieldNum)
		builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, messageName, node.Name, node.FieldNum))

		*fieldNum++
		childFieldNum := 1
		builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
		for _, child := range node.Children {
			p.writeProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ")
		}
		builder.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		protoType := p.mapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
	}
}

func (p *Parser) writeProtoFieldRecursive(builder *strings.Builder, node *TreeNode, fieldNum *int, indent string) {
	if node.Type == "message" || len(node.Children) > 0 {
		messageName := fmt.Sprintf("Message%d", *fieldNum)
		builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, messageName, node.Name, node.FieldNum))

		*fieldNum++
		childFieldNum := 1
		builder.WriteString(fmt.Sprintf("%smessage %s {\n", indent, messageName))
		for _, child := range node.Children {
			p.writeProtoFieldRecursive(builder, child, &childFieldNum, indent+"  ")
		}
		builder.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		protoType := p.mapTypeToProtoType(node.Type)
		if node.IsRepeated {
			builder.WriteString(fmt.Sprintf("%srepeated %s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s %s = %d;\n", indent, protoType, node.Name, node.FieldNum))
		}
	}
}

func (p *Parser) mapTypeToProtoType(ourType string) string {
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

func (p *Parser) treeToTextFormatWithNames(node *TreeNode) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	p.writeNodeToTextFormatWithNames(&result, node, 0)
	return result.String()
}

func (p *Parser) writeNodeToTextFormatWithNames(builder *strings.Builder, node *TreeNode, indent int) {
	if node.Name == "root" {
		for _, child := range node.Children {
			p.writeNodeToTextFormatWithNames(builder, child, indent)
		}
		return
	}

	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	if node.Type == "message" || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%s {\n", node.Name))
		for _, child := range node.Children {
			p.writeNodeToTextFormatWithNames(builder, child, indent+1)
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
