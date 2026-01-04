package ui

import (
	"fmt"
	"os"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ProtobufView создает UI для просмотра protobuf файлов
func ProtobufView(fyneApp fyne.App, parentWindow fyne.Window, browserTabs *BrowserTabs) fyne.CanvasObject {
	// Парсер protobuf
	parser, err := protobuf.NewParser()
	if err != nil {
		errorLabel := widget.NewLabel(fmt.Sprintf("Ошибка: %v", err))
		errorLabel.Wrapping = fyne.TextWrapWord
		return container.NewPadded(errorLabel)
	}

	// Дерево для отображения
	var currentTree *protobuf.TreeNode

	// Виджет дерева для визуального отображения
	treeWidget := CreateProtobufTree(nil)
	
	// Контейнер для дерева (будет обновляться)
	treeScrollContainer := container.NewScroll(treeWidget)

	// Получаем состояние диалогов
	dialogState := GetFileDialogState()

	// Объявляем переменные для кнопок и панели (определяются позже)
	var openBtn *widget.Button
	var applySchemaBtn *widget.Button
	var exportJSONBtn *widget.Button
	var buttonPanel fyne.CanvasObject

	// Кнопка открытия файла
	openBtn = widget.NewButton("Открыть бинарный protobuf файл", func() {
		fmt.Fprintf(os.Stdout, "[DEBUG] Кнопка открытия файла нажата\n")
		// Создаем диалог с сохранением позиции
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			fmt.Fprintf(os.Stdout, "[DEBUG] Callback диалога вызван, err=%v, reader=%v\n", err, reader != nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Ошибка диалога: %v\n", err)
				dialog.ShowError(err, parentWindow)
				return
			}
			if reader == nil {
				fmt.Fprintf(os.Stdout, "[DEBUG] Reader is nil, пользователь отменил выбор\n")
				return
			}
			defer reader.Close()
			fmt.Fprintf(os.Stdout, "[DEBUG] Файл выбран: %s\n", reader.URI().Path())

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
			fmt.Fprintf(os.Stdout, "[DEBUG] Дерево распарсено, узлов в root: %d\n", len(tree.Children))
			
			// ШАГ 1: Отображаем дерево в виде текста
			treeText := formatTree(tree, 0)
			fmt.Fprintf(os.Stdout, "[DEBUG] Форматированный текст дерева:\n%s\n", treeText)
			
			// Создаем текстовый виджет для отображения
			textView := widget.NewMultiLineEntry()
			textView.SetText(treeText)
			textView.Wrapping = fyne.TextWrapWord
			textView.Disable() // Только для чтения
			
			// Обновляем scroll контейнер с текстовым виджетом
			newScrollContainer := container.NewScroll(textView)
			treeScrollContainer = newScrollContainer
			
			// Обновляем Border контейнер
			newBorder := container.NewBorder(
				buttonPanel,
				nil,
				nil,
				nil,
				newScrollContainer,
			)
			// Обновляем контент вкладки через BrowserTabs
			if browserTabs != nil {
				fmt.Fprintf(os.Stdout, "[DEBUG] Вызов UpdateTabContent\n")
				browserTabs.UpdateTabContent(container.NewPadded(newBorder))
			} else {
				fmt.Fprintf(os.Stderr, "[ERROR] browserTabs is nil!\n")
			}
			fmt.Fprintf(os.Stdout, "[INFO] Protobuf файл успешно распарсен, дерево обновлено\n")
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
	applySchemaBtn = widget.NewButton("Применить proto схему", func() {
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
			// Обновляем виджет дерева
			adapter := NewProtobufTreeAdapter(tree)
			newTreeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
			newTreeWidget.OpenBranch("root")
			treeWidget = newTreeWidget
			// Обновляем scroll контейнер
			newScrollContainer := container.NewScroll(newTreeWidget)
			treeScrollContainer = newScrollContainer
			// Обновляем Border контейнер
			newBorder := container.NewBorder(
				buttonPanel,
				nil,
				nil,
				nil,
				newScrollContainer,
			)
			// Обновляем контент вкладки через BrowserTabs
			if browserTabs != nil {
				browserTabs.UpdateTabContent(container.NewPadded(newBorder))
			}
			fmt.Fprintf(os.Stdout, "[INFO] Схема успешно применена, дерево обновлено\n")
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
	exportJSONBtn = widget.NewButton("Экспорт в JSON", func() {
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
	buttonPanel = container.NewHBox(
		openBtn,
		applySchemaBtn,
		exportJSONBtn,
	)

	// Основной контейнер - создаем функцию для обновления
	createMainBorder := func() fyne.CanvasObject {
		return container.NewBorder(
			buttonPanel,           // верх - кнопки
			nil,                   // низ
			nil,                   // лево
			nil,                   // право
			treeScrollContainer,   // центр - дерево
		)
	}
	
	mainBorder := createMainBorder()
	content := mainBorder

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
	if node.Name == "root" {
		result += "Protobuf Root\n"
		result += strings.Repeat("=", 50) + "\n"
		if len(node.Children) == 0 {
			result += "(нет данных)\n"
			return result
		}
	} else {
		// Форматируем узел
		result += fmt.Sprintf("%s%s (field_%d, %s)", indentStr, node.Name, node.FieldNum, node.Type)
		if node.Value != nil {
			result += fmt.Sprintf(": %v", node.Value)
		}
		if node.IsRepeated {
			result += " [repeated]"
		}
		if len(node.Children) > 0 {
			result += fmt.Sprintf(" [%d children]", len(node.Children))
		}
		result += "\n"
	}

	// Рекурсивно форматируем дочерние узлы
	for _, child := range node.Children {
		result += formatTree(child, indent+1)
	}

	return result
}

