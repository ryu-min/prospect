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

type schemaFieldInfo struct {
	fieldNum   int
	fieldName  string
	fieldType  string
	isRepeated bool
	isRequired bool
	isOptional bool
}

type schemaMessageInfo struct {
	messageName string
	fields      []*schemaFieldInfo
	messages    map[string]*schemaMessageInfo
}

// ParseSchemaFile парсит proto файл и возвращает список имен сообщений верхнего уровня
func (p *Parser) ParseSchemaFile(schemaPath string) ([]string, error) {
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения схемы: %w", err)
	}

	_, topLevelMessages, err := p.parseProtoSchema(string(schemaContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга схемы: %w", err)
	}

	return topLevelMessages, nil
}

func (p *Parser) ApplySchema(tree *TreeNode, schemaPath string) (*TreeNode, error) {
	return p.ApplySchemaWithMessage(tree, schemaPath, "")
}

func (p *Parser) ApplySchemaWithMessage(tree *TreeNode, schemaPath string, messageName string) (*TreeNode, error) {
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения схемы: %w", err)
	}

	schema, topLevelMessages, err := p.parseProtoSchema(string(schemaContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга схемы: %w", err)
	}

	if len(schema) == 0 {
		return nil, fmt.Errorf("схема не содержит сообщений")
	}

	var rootMessage *schemaMessageInfo
	if messageName != "" {
		// Используем указанное сообщение
		rootMessage = schema[messageName]
		if rootMessage == nil {
			return nil, fmt.Errorf("сообщение '%s' не найдено в схеме", messageName)
		}
	} else {
		// Автоматический выбор: если сообщение одно, используем его, иначе ищем по приоритету
		if len(topLevelMessages) == 1 {
			rootMessage = schema[topLevelMessages[0]]
		} else {
			rootMessage = findRootMessage(schema, topLevelMessages)
		}
	}

	if rootMessage == nil {
		return nil, fmt.Errorf("не удалось найти корневое сообщение в схеме")
	}

	if err := p.validateSchema(tree, rootMessage); err != nil {
		return nil, err
	}

	p.applySchemaToTree(tree, rootMessage, schema)

	return tree, nil
}

func findRootMessage(messages map[string]*schemaMessageInfo, topLevelMessages []string) *schemaMessageInfo {
	// Сначала ищем сообщения с приоритетными именами среди сообщений верхнего уровня
	priorityNames := []string{"Message", "Root", "RootMessage"}
	for _, priorityName := range priorityNames {
		for _, topLevelName := range topLevelMessages {
			if topLevelName == priorityName {
				if msg, exists := messages[priorityName]; exists {
					return msg
				}
			}
		}
	}

	// Если не найдено, возвращаем первое сообщение верхнего уровня
	if len(topLevelMessages) > 0 {
		return messages[topLevelMessages[0]]
	}

	return nil
}

// parseProtoSchema парсит proto схему и возвращает:
// 1. Словарь всех сообщений (включая вложенные)
// 2. Список имен сообщений верхнего уровня
func (p *Parser) parseProtoSchema(content string) (map[string]*schemaMessageInfo, []string, error) {
	messages := make(map[string]*schemaMessageInfo)
	topLevelMessages := make([]string, 0)
	lines := strings.Split(content, "\n")

	var currentMessage *schemaMessageInfo
	var messageStack []*schemaMessageInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "syntax") || strings.HasPrefix(line, "package") || strings.HasPrefix(line, "import") {
			continue
		}

		if strings.HasPrefix(line, "message ") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			messageName := strings.TrimSuffix(strings.TrimSuffix(parts[1], "{"), " ")

			msg := &schemaMessageInfo{
				messageName: messageName,
				fields:      make([]*schemaFieldInfo, 0),
				messages:    make(map[string]*schemaMessageInfo),
			}

			if currentMessage != nil {
				// Это вложенное сообщение
				messageStack = append(messageStack, currentMessage)
				currentMessage.messages[messageName] = msg
				// Добавляем вложенное сообщение в общий словарь для доступа из других мест
				messages[messageName] = msg
			} else {
				// Это сообщение верхнего уровня
				messages[messageName] = msg
				topLevelMessages = append(topLevelMessages, messageName)
			}

			currentMessage = msg
			continue
		}

		if strings.HasPrefix(line, "}") {
			if len(messageStack) > 0 {
				currentMessage = messageStack[len(messageStack)-1]
				messageStack = messageStack[:len(messageStack)-1]
			} else {
				currentMessage = nil
			}
			continue
		}

		if currentMessage != nil {
			field := p.parseFieldLine(line)
			if field != nil {
				currentMessage.fields = append(currentMessage.fields, field)
			}
		}
	}

	return messages, topLevelMessages, nil
}

func (p *Parser) parseFieldLine(line string) *schemaFieldInfo {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "//") {
		return nil
	}

	parts := strings.Fields(line)
	if len(parts) < 4 {
		return nil
	}

	isRepeated := false
	isRequired := false
	isOptional := false

	i := 0
	if parts[i] == "repeated" {
		isRepeated = true
		i++
	} else if parts[i] == "required" {
		isRequired = true
		i++
	} else if parts[i] == "optional" {
		isOptional = true
		i++
	}

	if i >= len(parts) {
		return nil
	}

	fieldType := parts[i]
	i++

	if i >= len(parts) {
		return nil
	}

	fieldName := parts[i]
	i++

	if i >= len(parts) {
		return nil
	}

	if parts[i] != "=" {
		return nil
	}
	i++

	if i >= len(parts) {
		return nil
	}

	fieldNumStr := strings.TrimSuffix(parts[i], ";")
	fieldNumStr = strings.TrimSpace(fieldNumStr)
	fieldNum, err := strconv.Atoi(fieldNumStr)
	if err != nil {
		return nil
	}

	if !isRequired && !isOptional && !isRepeated {
		isOptional = true
	}

	return &schemaFieldInfo{
		fieldNum:   fieldNum,
		fieldName:  fieldName,
		fieldType:  fieldType,
		isRepeated: isRepeated,
		isRequired: isRequired,
		isOptional: isOptional,
	}
}

func (p *Parser) validateSchema(tree *TreeNode, schema *schemaMessageInfo) error {
	treeFields := make(map[int]bool)
	p.collectFieldNums(tree, treeFields)

	for _, field := range schema.fields {
		if field.isRequired && !treeFields[field.fieldNum] {
			return fmt.Errorf("отсутствует обязательное поле '%s' (номер поля: %d)", field.fieldName, field.fieldNum)
		}
	}

	return nil
}

func (p *Parser) collectFieldNums(node *TreeNode, fieldNums map[int]bool) {
	if node.FieldNum > 0 {
		fieldNums[node.FieldNum] = true
	}

	for _, child := range node.Children {
		p.collectFieldNums(child, fieldNums)
	}
}

func (p *Parser) applySchemaToTree(tree *TreeNode, schema *schemaMessageInfo, allMessages map[string]*schemaMessageInfo) {
	fieldMap := make(map[int]*schemaFieldInfo)
	for _, field := range schema.fields {
		fieldMap[field.fieldNum] = field
	}

	for _, child := range tree.Children {
		if fieldInfo, ok := fieldMap[child.FieldNum]; ok {
			oldType := child.Type
			child.Name = fieldInfo.fieldName

			// Если тип поля - это тип сообщения, устанавливаем тип поля сразу
			if p.isMessageTypeName(fieldInfo.fieldType) {
				child.Type = fieldInfo.fieldType
			} else {
				// Для не-сообщений применяем обычное преобразование типа
				newType := p.mapProtoTypeToUIType(fieldInfo.fieldType)
				if p.canConvertType(oldType, newType, child.Value) {
					child.Type = newType
					child.Value = p.convertValue(child.Value, oldType, newType)
				}
			}

			child.IsRepeated = fieldInfo.isRepeated

			// Применяем схему к вложенным сообщениям
			if p.isMessageTypeName(fieldInfo.fieldType) {
				// Ищем схему вложенного сообщения сначала в schema.messages (локальные вложенные сообщения),
				// затем в allMessages (все сообщения)
				if nestedSchema, ok := schema.messages[fieldInfo.fieldType]; ok {
					p.applySchemaToTree(child, nestedSchema, allMessages)
				} else if nestedSchema, ok := allMessages[fieldInfo.fieldType]; ok {
					p.applySchemaToTree(child, nestedSchema, allMessages)
				}
			}
		}
	}
}

func (p *Parser) mapProtoTypeToUIType(protoType string) string {
	switch protoType {
	case "string", "bytes":
		return "string"
	case "int32", "sint32", "sfixed32":
		return "int32"
	case "int64", "sint64", "sfixed64":
		return "int64"
	case "uint32", "fixed32":
		return "uint32"
	case "uint64", "fixed64":
		return "uint64"
	case "bool":
		return "bool"
	case "float":
		return "float"
	case "double":
		return "double"
	default:
		// Если это не базовый тип, проверяем, является ли это типом сообщения
		// Типы сообщений могут быть как "Message1", так и "message_1"
		if p.isMessageTypeName(protoType) {
			return protoType
		}
		return "string"
	}
}

func (p *Parser) isMessageTypeName(typeName string) bool {
	// Проверяем, является ли тип именем сообщения (не базовым типом)
	basicTypes := map[string]bool{
		"string": true, "bytes": true,
		"int32": true, "sint32": true, "sfixed32": true,
		"int64": true, "sint64": true, "sfixed64": true,
		"uint32": true, "fixed32": true,
		"uint64": true, "fixed64": true,
		"bool":  true,
		"float": true, "double": true,
	}

	if basicTypes[typeName] {
		return false
	}

	// Если это не базовый тип, то это скорее всего тип сообщения
	return true
}

func (p *Parser) canConvertType(oldType, newType string, value interface{}) bool {
	if oldType == newType {
		return true
	}

	if isMessageType(oldType) && isMessageType(newType) {
		return true
	}

	if isMessageType(oldType) || isMessageType(newType) {
		return false
	}

	if value == nil {
		return true
	}

	valueStr := fmt.Sprintf("%v", value)

	if oldType == "string" && newType == "string" {
		return true
	}

	if (oldType == "int32" || oldType == "int64" || oldType == "uint32" || oldType == "uint64" || oldType == "sint32" || oldType == "sint64") &&
		(newType == "int32" || newType == "int64" || newType == "uint32" || newType == "uint64" || newType == "sint32" || newType == "sint64") {
		return p.isCompatibleIntegerType(oldType, newType, valueStr)
	}

	if (oldType == "float" || oldType == "double") && (newType == "float" || newType == "double") {
		return true
	}

	if (oldType == "int32" || oldType == "int64" || oldType == "uint32" || oldType == "uint64" || oldType == "sint32" || oldType == "sint64") &&
		(newType == "float" || newType == "double") {
		return true
	}

	if (oldType == "float" || oldType == "double") &&
		(newType == "int32" || newType == "int64" || newType == "uint32" || newType == "uint64" || newType == "sint32" || newType == "sint64") {
		if _, err := strconv.ParseFloat(valueStr, 64); err == nil {
			f, _ := strconv.ParseFloat(valueStr, 64)
			return float64(int64(f)) == f
		}
		return false
	}

	if oldType == "bool" && (newType == "int32" || newType == "int64") {
		return true
	}

	if (oldType == "int32" || oldType == "int64") && newType == "bool" {
		return valueStr == "0" || valueStr == "1" || valueStr == "true" || valueStr == "false"
	}

	return false
}

func (p *Parser) isCompatibleIntegerType(oldType, newType, valueStr string) bool {
	if oldType == newType {
		return true
	}

	val, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		if _, uerr := strconv.ParseUint(valueStr, 10, 64); uerr == nil {
			if strings.Contains(newType, "uint") || strings.Contains(newType, "int") {
				return true
			}
			return false
		}
		return false
	}

	if strings.Contains(newType, "uint") {
		return val >= 0
	}

	if newType == "int32" || newType == "sint32" {
		return val >= -2147483648 && val <= 2147483647
	}

	if newType == "uint32" {
		return val >= 0 && val <= 4294967295
	}

	return true
}

func (p *Parser) convertValue(value interface{}, oldType, newType string) interface{} {
	if value == nil {
		return nil
	}

	if oldType == newType {
		return value
	}

	valueStr := fmt.Sprintf("%v", value)

	if (oldType == "int32" || oldType == "int64" || oldType == "uint32" || oldType == "uint64" || oldType == "sint32" || oldType == "sint64") &&
		(newType == "float" || newType == "double") {
		if val, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return fmt.Sprintf("%d.0", val)
		}
		if val, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
			return fmt.Sprintf("%d.0", val)
		}
		return valueStr
	}

	if (oldType == "float" || oldType == "double") &&
		(newType == "int32" || newType == "int64" || newType == "uint32" || newType == "uint64" || newType == "sint32" || newType == "sint64") {
		if f, err := strconv.ParseFloat(valueStr, 64); err == nil {
			if newType == "uint32" || newType == "uint64" {
				if f >= 0 {
					return fmt.Sprintf("%.0f", f)
				}
				return "0"
			}
			return fmt.Sprintf("%.0f", f)
		}
		return valueStr
	}

	if oldType == "bool" && (newType == "int32" || newType == "int64") {
		if b, ok := value.(bool); ok {
			if b {
				return "1"
			}
			return "0"
		}
		if valueStr == "true" || valueStr == "1" {
			return "1"
		}
		return "0"
	}

	if (oldType == "int32" || oldType == "int64") && newType == "bool" {
		if valueStr == "1" {
			return true
		}
		return false
	}

	return value
}

// GetProtocPath возвращает путь к protoc (для использования в тестах и других компонентах)
func (p *Parser) GetProtocPath() string {
	return p.protocPath
}

func (n *TreeNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}
