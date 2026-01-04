package protobuf

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	fmt.Fprintf(os.Stdout, "[INFO] protoc найден: %s\n", strings.TrimSpace(string(output)))
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
	fmt.Fprintf(os.Stdout, "[DEBUG] Вывод protoc:\n%s\n", outputStr)
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
	stackIndents := []int{0} // Отслеживаем отступы для каждого уровня стека

	for _, line := range lines {
		originalLine := line
		line = strings.TrimRight(line, " \t")
		if line == "" {
			continue
		}

		indent := getIndentLevel(originalLine)
		trimmedLine := strings.TrimLeft(line, " \t")

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
			// Если отступ меньше текущего уровня стека, выходим из вложенных сообщений
			for len(stack) > 1 && len(stackIndents) > 1 {
				lastIndent := stackIndents[len(stackIndents)-1]
				if indent <= lastIndent {
					stack = stack[:len(stack)-1]
					stackIndents = stackIndents[:len(stackIndents)-1]
				} else {
					break
				}
			}
			if len(stack) > 0 {
				stack[len(stack)-1].AddChild(node)
			}
			stack = append(stack, node)
			stackIndents = append(stackIndents, indent)
			fmt.Fprintf(os.Stdout, "[DEBUG] Открыто сообщение: %s (indent=%d, stackSize=%d)\n",
				node.Name, indent, len(stack))
			continue
		}

		// Если строка == "}", это конец вложенного сообщения
		if trimmedLine == "}" {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
				stackIndents = stackIndents[:len(stackIndents)-1]
			}
			continue
		}

		// Если отступ меньше текущего уровня стека, выходим из вложенных сообщений
		// Нужно найти правильного родителя на основе отступа
		for len(stack) > 1 && len(stackIndents) > 1 {
			lastIndent := stackIndents[len(stackIndents)-1]
			// Если текущий отступ меньше или равен отступу последнего элемента стека,
			// значит мы вышли из этого сообщения - удаляем его из стека
			if indent <= lastIndent {
				stack = stack[:len(stack)-1]
				stackIndents = stackIndents[:len(stackIndents)-1]
			} else {
				// Отступ больше - мы все еще внутри сообщения
				break
			}
		}

		// Парсим строку
		node := p.parseLine(trimmedLine)
		if node != nil {
			// Добавляем в текущий узел стека (после корректировки стека)
			if len(stack) > 0 {
				parentName := stack[len(stack)-1].Name
				stack[len(stack)-1].AddChild(node)
				parentIndent := 0
				if len(stackIndents) > 0 {
					parentIndent = stackIndents[len(stackIndents)-1]
				}
				fmt.Fprintf(os.Stdout, "[DEBUG] Добавлен узел: %s (field_%d, %s) в %s (indent=%d, parentIndent=%d, stackSize=%d)\n",
					node.Name, node.FieldNum, node.Type, parentName, indent, parentIndent, len(stack))
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

// parseLine парсит одну строку вывода protoc
func (p *Parser) parseLine(line string) *TreeNode {
	// Формат protoc --decode_raw:
	//   field_number: value
	//   field_number {
	//     nested_field: value
	//   }
	// Примеры:
	//   1: "hello"
	//   2: 42
	//   3: 1
	//   4 {
	//     5: "nested"
	//   }

	line = strings.TrimSpace(line)
	if line == "" || line == "{" || line == "}" {
		return nil
	}

	// Проверяем, является ли это началом вложенного сообщения
	if strings.HasSuffix(line, "{") {
		fieldPart := strings.TrimSpace(strings.TrimSuffix(line, "{"))
		fieldNum := parseInt(fieldPart)
		return &TreeNode{
			Name:       fmt.Sprintf("field_%d", fieldNum),
			Type:       "message",
			FieldNum:   fieldNum,
			Children:   make([]*TreeNode, 0),
			IsRepeated: false,
		}
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
