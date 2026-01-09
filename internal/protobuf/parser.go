package protobuf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
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

	renumberMessages(root)
	return root, nil
}

func renumberMessages(root *TreeNode) {
	messageCounter := 1
	renumberMessagesRecursive(root, &messageCounter)
}

func renumberMessagesRecursive(node *TreeNode, counter *int) {
	if node == nil {
		return
	}

	if isMessageType(node.Type) && node.Name != "root" {
		node.Type = fmt.Sprintf("message_%d", *counter)
		*counter++
	}

	for _, child := range node.Children {
		renumberMessagesRecursive(child, counter)
	}
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
	} else if isHexFloat(valueStr) {
		node.Type = "double"
		decimalValue := convertHexFloatToDecimal(valueStr)
		node.Value = decimalValue
	} else if isFloat(valueStr) {
		node.Type = "double"
		node.Value = valueStr
	} else if isInteger(valueStr) {
		normalizedValue := parseSignedNumber(valueStr)
		if normalizedValue == "0" || normalizedValue == "1" {
			node.Type = "bool"
			if normalizedValue == "1" {
				node.Value = true
			} else {
				node.Value = false
			}
		} else {
			node.Type = "int64"
			node.Value = normalizedValue
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
	var num int64
	_, err := fmt.Sscanf(s, "%d", &num)
	if err == nil {
		return true
	}
	var unum uint64
	_, err = fmt.Sscanf(s, "%d", &unum)
	if err == nil {
		return true
	}
	var fnum float64
	_, err = fmt.Sscanf(s, "%f", &fnum)
	return err == nil
}

func isInteger(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return true
	}
	_, err = strconv.ParseUint(s, 10, 64)
	return err == nil
}

func isFloat(s string) bool {
	if isInteger(s) {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isHexFloat(s string) bool {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		var num uint64
		_, err := fmt.Sscanf(s, "%x", &num)
		return err == nil
	}
	return false
}

func convertHexFloatToDecimal(hexStr string) string {
	hexStr = strings.TrimPrefix(strings.TrimPrefix(hexStr, "0x"), "0X")
	var bits uint64
	_, err := fmt.Sscanf(hexStr, "%x", &bits)
	if err != nil {
		return hexStr
	}
	
	floatValue := math.Float64frombits(bits)
	return strconv.FormatFloat(floatValue, 'g', -1, 64)
}

func parseSignedNumber(s string) string {
	var unum uint64
	_, err := fmt.Sscanf(s, "%d", &unum)
	if err != nil {
		return s
	}
	
	const maxInt64 = uint64(1) << 63
	if unum >= maxInt64 {
		signedNum := int64(unum)
		return fmt.Sprintf("%d", signedNum)
	}
	
	return s
}

func (p *Parser) ApplySchema(tree *TreeNode, schemaPath string) (*TreeNode, error) {
	return tree, nil
}

// GetProtocPath возвращает путь к protoc (для использования в тестах и других компонентах)
func (p *Parser) GetProtocPath() string {
	return p.protocPath
}

func (n *TreeNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}
