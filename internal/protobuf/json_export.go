package protobuf

import (
	"encoding/json"
	"fmt"
)

// TreeNodeToJSON конвертирует TreeNode в JSON объект
// Ключи - имена полей из ноды (используется Name, которое содержит имя из схемы если она была применена)
// Значения - значения полей, для message типов - вложенные JSON объекты
func TreeNodeToJSON(node *TreeNode) (map[string]interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}

	result := make(map[string]interface{})

	// Обрабатываем дочерние элементы (поля сообщения)
	for _, child := range node.Children {
		fieldName := child.Name
		if fieldName == "" {
			fieldName = fmt.Sprintf("field_%d", child.FieldNum)
		}

		var value interface{}

		// Если это message тип (есть дочерние элементы или тип message)
		if child.IsMessage() {
			// Рекурсивно конвертируем вложенное сообщение
			nestedJSON, err := TreeNodeToJSON(child)
			if err != nil {
				return nil, fmt.Errorf("error converting nested message for field %s: %w", fieldName, err)
			}
			value = nestedJSON
		} else {
			// Простое значение
			value = child.Value
		}

		// Если поле повторяющееся (repeated), создаем массив
		if child.IsRepeated {
			// Проверяем, есть ли уже массив для этого поля
			if existing, exists := result[fieldName]; exists {
				if arr, ok := existing.([]interface{}); ok {
					result[fieldName] = append(arr, value)
				} else {
					// Если уже есть значение, но не массив, создаем массив из существующего и нового значения
					result[fieldName] = []interface{}{existing, value}
				}
			} else {
				result[fieldName] = []interface{}{value}
			}
		} else {
			// Обычное поле
			result[fieldName] = value
		}
	}

	return result, nil
}

// TreeNodeToJSONString конвертирует TreeNode в JSON строку с форматированием
func TreeNodeToJSONString(node *TreeNode) (string, error) {
	jsonObj, err := TreeNodeToJSON(node)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

