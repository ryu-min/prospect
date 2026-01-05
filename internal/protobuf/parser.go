package protobuf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Parser парсит бинарные protobuf данные
type Parser struct {
	protocPath string
}

// NewParser создает новый парсер protobuf
func NewParser() (*Parser, error) {
	protocPath, err := findProtoc()
	if err != nil {
		return nil, fmt.Errorf("protoc не найден: %w", err)
	}

	return &Parser{
		protocPath: protocPath,
	}, nil
}

// findProtoc ищет путь к protoc в системе
func findProtoc() (string, error) {
	// Сначала проверяем, доступен ли protoc в PATH
	cmd := exec.Command("protoc", "--version")
	if err := cmd.Run(); err == nil {
		return "protoc", nil
	}

	// Проверяем возможные пути установки
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

// CheckProtoc проверяет, установлен ли protoc
func CheckProtoc() error {
	parser, err := NewParser()
	if err != nil {
		return err
	}

	// Проверяем, что protoc работает
	cmd := exec.Command(parser.protocPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("protoc не работает: %w", err)
	}

	log.Printf("protoc найден: %s", strings.TrimSpace(string(output)))
	return nil
}

// ParseRaw декодирует бинарные protobuf данные в дерево
func (p *Parser) ParseRaw(data []byte) (*TreeNode, error) {
	// Используем protoc --decode raw для декодирования
	cmd := exec.Command(p.protocPath, "--decode_raw")
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования protobuf: %w", err)
	}

	// Парсим вывод protoc в дерево
	outputStr := string(output)
	return p.parseProtocOutput(outputStr)
}

// parseProtocOutput парсит текстовый вывод protoc в дерево
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
	stackIndents := []int{-1} // root имеет отступ -1, чтобы любое поле с отступом >= 0 было его ребенком

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		indent := getIndentLevel(originalLine)
		trimmedLine := line

		// Если строка == "}", это конец вложенного сообщения
		if trimmedLine == "}" {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
				stackIndents = stackIndents[:len(stackIndents)-1]
			}
			continue
		}

		// Если отступ меньше или равен текущему уровню стека, выходим из вложенных сообщений
		// Мы должны найти родителя, чей отступ МЕНЬШЕ текущего
		for len(stack) > 1 && indent <= stackIndents[len(stackIndents)-1] {
			stack = stack[:len(stack)-1]
			stackIndents = stackIndents[:len(stackIndents)-1]
		}

		// Если строка заканчивается на "{", это начало вложенного сообщения
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

		// Парсим строку
		node := p.parseLine(trimmedLine)
		if node != nil {
			// Добавляем в текущий узел стека (после корректировки стека)
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.AddChild(node)
			}
		}
	}

	return root, nil
}

// getIndentLevel возвращает уровень отступа строки
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

// parseLine парсит одну строку вывода protoc (только пары field: value)
func (p *Parser) parseLine(line string) *TreeNode {
	line = strings.TrimSpace(line)
	if line == "" || line == "{" || line == "}" || strings.HasSuffix(line, "{") {
		return nil
	}

	// Парсим поле со значением
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	fieldNum := parseInt(strings.TrimSpace(parts[0]))
	valueStr := strings.TrimSpace(parts[1])

	// Определяем тип и значение
	node := &TreeNode{
		FieldNum: fieldNum,
		Name:     fmt.Sprintf("field_%d", fieldNum),
	}

	// Парсим значение
	if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
		// Строка
		node.Type = "string"
		node.Value = strings.Trim(valueStr, "\"")
	} else if isNumeric(valueStr) {
		// Число (может быть int32, int64, uint32, uint64, bool и т.д.)
		node.Type = "number"
		node.Value = valueStr
		// Для bool: 0 = false, 1 = true
		if valueStr == "0" || valueStr == "1" {
			node.Type = "bool"
			if valueStr == "1" {
				node.Value = true
			} else {
				node.Value = false
			}
		}
	} else {
		// Неизвестный тип, сохраняем как есть
		node.Type = "unknown"
		node.Value = valueStr
	}

	return node
}

// parseInt парсит строку в число
func parseInt(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

// isNumeric проверяет, является ли строка числом
func isNumeric(s string) bool {
	_, err := fmt.Sscanf(s, "%d", new(int))
	return err == nil
}

// ApplySchema применяет proto схему к дереву (заглушка для будущей реализации)
func (p *Parser) ApplySchema(tree *TreeNode, schemaPath string) (*TreeNode, error) {
	// TODO: Реализовать применение схемы через protoc --decode
	// Это будет использоваться для декодирования с известной схемой
	return tree, nil
}

// ToJSON конвертирует дерево в JSON для отладки
func (n *TreeNode) ToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}

// SerializeRaw сериализует дерево обратно в бинарный формат protobuf
func (p *Parser) SerializeRaw(tree *TreeNode) ([]byte, error) {
	// Преобразуем дерево в текстовый формат protoc
	textFormat := p.treeToTextFormat(tree)

	// Используем protoc --encode_raw для кодирования
	cmd := exec.Command(p.protocPath, "--encode_raw")
	cmd.Stdin = strings.NewReader(textFormat)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error encoding protobuf: %w", err)
	}

	return output, nil
}

// treeToTextFormat преобразует дерево в текстовый формат protoc
func (p *Parser) treeToTextFormat(node *TreeNode) string {
	if node == nil {
		return ""
	}

	var result strings.Builder
	p.writeNodeToTextFormat(&result, node, 0)
	return result.String()
}

// writeNodeToTextFormat рекурсивно записывает узел в текстовый формат
func (p *Parser) writeNodeToTextFormat(builder *strings.Builder, node *TreeNode, indent int) {
	if node.Name == "root" {
		// Пропускаем root, обрабатываем только детей
		for _, child := range node.Children {
			p.writeNodeToTextFormat(builder, child, indent)
		}
		return
	}

	// Добавляем отступ
	for i := 0; i < indent; i++ {
		builder.WriteString("  ")
	}

	// Если это сообщение, открываем блок
	if node.Type == "message" || len(node.Children) > 0 {
		builder.WriteString(fmt.Sprintf("%d {\n", node.FieldNum))
		// Рекурсивно обрабатываем детей
		for _, child := range node.Children {
			p.writeNodeToTextFormat(builder, child, indent+1)
		}
		// Закрываем блок
		for i := 0; i < indent; i++ {
			builder.WriteString("  ")
		}
		builder.WriteString("}\n")
	} else {
		// Примитивное значение
		builder.WriteString(fmt.Sprintf("%d: ", node.FieldNum))
		if node.Value != nil {
			// Обрабатываем в зависимости от типа поля
			if node.Type == "string" {
				// Строка - всегда в кавычках
				builder.WriteString(fmt.Sprintf("\"%s\"", fmt.Sprintf("%v", node.Value)))
			} else if node.Type == "bool" {
				// Bool - true/false
				if v, ok := node.Value.(bool); ok {
					if v {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				} else {
					// Если сохранено как строка
					valueStr := fmt.Sprintf("%v", node.Value)
					if valueStr == "true" || valueStr == "1" {
						builder.WriteString("true")
					} else {
						builder.WriteString("false")
					}
				}
			} else if node.Type == "number" {
				// Число - без кавычек
				builder.WriteString(fmt.Sprintf("%v", node.Value))
			} else {
				// Неизвестный тип - пытаемся определить
				switch v := node.Value.(type) {
				case string:
					// Проверяем, является ли это числом
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
