package ui

import (
	"fmt"
	"os"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ProtobufView создает UI для просмотра protobuf файлов
func ProtobufView(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	// Парсер protobuf
	parser, err := protobuf.NewParser()
	if err != nil {
		errorLabel := widget.NewLabel(fmt.Sprintf("Ошибка: %v", err))
		errorLabel.Wrapping = fyne.TextWrapWord
		return container.NewPadded(errorLabel)
	}

	// Дерево для отображения
	var currentTree *protobuf.TreeNode

	// Текстовое представление дерева
	treeText := widget.NewMultiLineEntry()
	treeText.Wrapping = fyne.TextWrapWord
	treeText.Disable() // Делаем только для чтения

	// Получаем состояние диалогов
	dialogState := GetFileDialogState()

	// Кнопка открытия файла
	openBtn := widget.NewButton("Открыть бинарный protobuf файл", func() {
		// Создаем диалог с сохранением позиции
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Сохраняем последнюю директорию
			dialogState.SetLastOpenDir(reader.URI())

			// Читаем данные
			data := make([]byte, 0)
			buf := make([]byte, 4096)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					data = append(data, buf[:n]...)
				}
				if err != nil {
					break
				}
			}

			// Парсим protobuf
			fmt.Fprintf(os.Stdout, "[INFO] Парсинг protobuf файла: %s\n", reader.URI().Path())
			tree, err := parser.ParseRaw(data)
			if err != nil {
				dialog.ShowError(fmt.Errorf("ошибка парсинга: %w", err), parentWindow)
				return
			}

			currentTree = tree
			// Отображаем дерево
			treeStr := formatTree(tree, 0)
			fmt.Fprintf(os.Stdout, "[DEBUG] Отформатированное дерево:\n%s\n", treeStr)
			treeText.SetText(treeStr)
			treeText.Refresh() // Принудительно обновляем виджет
			fmt.Fprintf(os.Stdout, "[INFO] Protobuf файл успешно распарсен\n")
		}, parentWindow)

		// Устанавливаем начальную директорию
		if lastDir := dialogState.GetLastOpenDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		// Показываем диалог
		fileDialog.Resize(dialogState.GetDialogSize())
		fileDialog.Show()
	})

	// Кнопка применения схемы
	applySchemaBtn := widget.NewButton("Применить proto схему", func() {
		if currentTree == nil {
			dialog.ShowInformation("Информация", "Сначала откройте protobuf файл", parentWindow)
			return
		}

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Сохраняем последнюю директорию для схем
			dialogState.SetLastSchemaDir(reader.URI())

			schemaPath := reader.URI().Path()
			fmt.Fprintf(os.Stdout, "[INFO] Применение схемы: %s\n", schemaPath)

			// TODO: Реализовать применение схемы
			tree, err := parser.ApplySchema(currentTree, schemaPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("ошибка применения схемы: %w", err), parentWindow)
				return
			}

			currentTree = tree
			treeText.SetText(formatTree(tree, 0))
			fmt.Fprintf(os.Stdout, "[INFO] Схема успешно применена\n")
		}, parentWindow)

		// Устанавливаем начальную директорию
		if lastDir := dialogState.GetLastSchemaDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		// Показываем диалог
		fileDialog.Resize(dialogState.GetDialogSize())
		fileDialog.Show()
	})

	// Кнопка экспорта в JSON
	exportJSONBtn := widget.NewButton("Экспорт в JSON", func() {
		if currentTree == nil {
			dialog.ShowInformation("Информация", "Сначала откройте protobuf файл", parentWindow)
			return
		}

		jsonData, err := currentTree.ToJSON()
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка экспорта: %w", err), parentWindow)
			return
		}

		fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parentWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			// Сохраняем последнюю директорию для сохранения
			dialogState.SetLastSaveDir(writer.URI())

			if _, err := writer.Write(jsonData); err != nil {
				dialog.ShowError(fmt.Errorf("ошибка записи: %w", err), parentWindow)
				return
			}

			dialog.ShowInformation("Успех", "JSON файл сохранен", parentWindow)
		}, parentWindow)

		// Устанавливаем начальную директорию
		if lastDir := dialogState.GetLastSaveDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		// Показываем диалог
		fileDialog.Resize(dialogState.GetDialogSize())
		fileDialog.Show()
	})

	// Панель кнопок
	buttonPanel := container.NewHBox(
		openBtn,
		applySchemaBtn,
		exportJSONBtn,
	)

	// Основной контейнер
	content := container.NewBorder(
		buttonPanel,                    // верх - кнопки
		nil,                            // низ
		nil,                            // лево
		nil,                            // право
		container.NewScroll(treeText),  // центр - дерево
	)

	return container.NewPadded(content)
}

// formatTree форматирует дерево для отображения
func formatTree(node *protobuf.TreeNode, indent int) string {
	if node == nil {
		return "(пустое дерево)\n"
	}

	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	result := ""
	if node.Name != "root" {
		result += fmt.Sprintf("%s%s (field_%d, %s)", indentStr, node.Name, node.FieldNum, node.Type)
		if node.Value != nil {
			result += fmt.Sprintf(": %v", node.Value)
		}
		if node.IsRepeated {
			result += " [repeated]"
		}
		result += "\n"
	}

	// Если это root и нет детей, показываем сообщение
	if node.Name == "root" && len(node.Children) == 0 {
		return "(нет данных)\n"
	}

	for _, child := range node.Children {
		result += formatTree(child, indent)
	}

	return result
}

