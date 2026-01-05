package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// ProtobufView создает UI для просмотра protobuf файлов
func ProtobufView(fyneApp fyne.App, parentWindow fyne.Window, browserTabs *BrowserTabs) fyne.CanvasObject {
	// Парсер protobuf
	parser, err := protobuf.NewParser()
	if err != nil {
		errorLabel := widget.NewLabel(fmt.Sprintf("Error: %v", err))
		errorLabel.Wrapping = fyne.TextWrapWord
		return container.NewPadded(errorLabel)
	}

	// Дерево для отображения
	var currentTree *protobuf.TreeNode
	var currentFilePath string // Путь к текущему открытому файлу

	// Виджет дерева для визуального отображения
	treeWidget := CreateProtobufTree(nil)

	// Контейнер для дерева (будет обновляться)
	treeScrollContainer := container.NewScroll(treeWidget)

	// Получаем состояние диалогов
	dialogState := GetFileDialogState()

	// Объявляем переменные для кнопок и панели (определяются позже)
	var openBtn *widget.Button
	var saveBtn *widget.Button
	var applySchemaBtn *widget.Button
	var exportJSONBtn *widget.Button
	var buttonPanel fyne.CanvasObject

	// Кнопка открытия файла
	openBtn = widget.NewButton("Open binary protobuf file", func() {
		// Создаем диалог с сохранением позиции
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				log.Printf("Dialog error: %v", err)
				dialog.ShowError(err, parentWindow)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Сохраняем путь к файлу
			currentFilePath = reader.URI().Path()

			// Обновляем название вкладки с именем файла
			if browserTabs != nil {
				fileName := filepath.Base(currentFilePath)
				browserTabs.UpdateTabTitle(fileName)
			}

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
			log.Printf("Parsing protobuf file: %s", reader.URI().Path())

			tree, err := parser.ParseRaw(data)
			if err != nil {
				dialog.ShowError(fmt.Errorf("parsing error: %w", err), parentWindow)
				return
			}

			currentTree = tree

			// ШАГ 3: Отображаем дерево в виде дерева (widget.Tree)
			adapter := NewProtobufTreeAdapter(tree)
			adapter.SetWindow(parentWindow) // Устанавливаем окно для диалогов

			// Создаем виджет дерева
			newTreeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)

			// Проверяем, что root имеет детей и открываем его
			// В Fyne widget.Tree использует пустую строку "" для root
			rootChildren := adapter.ChildUIDs("")

			// ВАЖНО: Открываем root ДО добавления в контейнер (используем пустую строку)
			if len(rootChildren) > 0 {
				newTreeWidget.OpenBranch("")
			}

			// Обновляем виджет дерева
			newTreeWidget.Refresh()

			treeWidget = newTreeWidget

			// Обновляем scroll контейнер с деревом
			newScrollContainer := container.NewScroll(newTreeWidget)
			newScrollContainer.Refresh()
			treeScrollContainer = newScrollContainer

			// ВАЖНО: Принудительно запрашиваем данные для root
			_ = adapter.ChildUIDs("root")
			_ = adapter.IsBranch("root")

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
			} else {
				log.Printf("Error: browserTabs is nil")
			}
			log.Printf("Protobuf file parsed successfully, tree updated")
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
	applySchemaBtn = widget.NewButton("Apply proto schema", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a protobuf file first", parentWindow)
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
			log.Printf("Applying schema: %s", schemaPath)

			// TODO: Реализовать применение схемы
			tree, err := parser.ApplySchema(currentTree, schemaPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("error applying schema: %w", err), parentWindow)
				return
			}

			currentTree = tree
			// Обновляем виджет дерева
			adapter := NewProtobufTreeAdapter(tree)
			adapter.SetWindow(parentWindow) // Устанавливаем окно для диалогов
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
			log.Printf("Schema applied successfully, tree updated")
		}, parentWindow)

		// Устанавливаем начальную директорию
		if lastDir := dialogState.GetLastSchemaDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		// Показываем диалог
		fileDialog.Resize(dialogState.GetDialogSize())
		fileDialog.Show()
	})

	// Кнопка сохранения
	saveBtn = widget.NewButton("Save", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a protobuf file first", parentWindow)
			return
		}

		// Сериализуем дерево в бинарный формат
		binaryData, err := parser.SerializeRaw(currentTree)
		if err != nil {
			dialog.ShowError(fmt.Errorf("serialization error: %w", err), parentWindow)
			return
		}

		// Создаем диалог сохранения
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				log.Printf("Save dialog error: %v", err)
				dialog.ShowError(err, parentWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			// Сохраняем путь к файлу
			currentFilePath = writer.URI().Path()

			// Сохраняем последнюю директорию
			dialogState.SetLastSaveDir(writer.URI())

			// Записываем данные
			if _, err := writer.Write(binaryData); err != nil {
				dialog.ShowError(fmt.Errorf("write error: %w", err), parentWindow)
				return
			}

			dialog.ShowInformation("Success", "Protobuf file saved", parentWindow)
			log.Printf("Protobuf file saved: %s", currentFilePath)
		}, parentWindow)

		// Устанавливаем начальную директорию
		if lastDir := dialogState.GetLastSaveDir(); lastDir != nil {
			saveDialog.SetLocation(lastDir)
		} else if currentFilePath != "" {
			// Используем директорию текущего файла
			dirPath := filepath.Dir(currentFilePath)
			uri := storage.NewFileURI(dirPath)
			if listableURI, err := storage.ListerForURI(uri); err == nil {
				saveDialog.SetLocation(listableURI)
			}
		}

		// Показываем диалог
		saveDialog.Resize(dialogState.GetDialogSize())
		saveDialog.Show()
	})

	// Кнопка экспорта в JSON
	exportJSONBtn = widget.NewButton("Export to JSON", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a protobuf file first", parentWindow)
			return
		}

		jsonData, err := currentTree.ToJSON()
		if err != nil {
			dialog.ShowError(fmt.Errorf("export error: %w", err), parentWindow)
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
				dialog.ShowError(fmt.Errorf("write error: %w", err), parentWindow)
				return
			}

			dialog.ShowInformation("Success", "JSON file saved", parentWindow)
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
		saveBtn,
		applySchemaBtn,
		exportJSONBtn,
	)

	// Основной контейнер - создаем функцию для обновления
	createMainBorder := func() fyne.CanvasObject {
		return container.NewBorder(
			buttonPanel,         // верх - кнопки
			nil,                 // низ
			nil,                 // лево
			nil,                 // право
			treeScrollContainer, // центр - дерево
		)
	}

	mainBorder := createMainBorder()
	content := mainBorder

	return container.NewPadded(content)
}

// formatTree форматирует дерево для отображения
func formatTree(node *protobuf.TreeNode, indent int) string {
	if node == nil {
		return "(empty tree)\n"
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
			result += "(no data)\n"
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

// TableRow представляет строку таблицы
type TableRow struct {
	FieldName string
	FieldNum  int
	Type      string
	Value     string
	Children  int
}

// buildTableData преобразует дерево в табличные данные
func buildTableData(node *protobuf.TreeNode) []TableRow {
	var rows []TableRow

	var traverse func(*protobuf.TreeNode, int)
	traverse = func(n *protobuf.TreeNode, level int) {
		if n == nil {
			return
		}

		// Пропускаем root узел, но обрабатываем его детей
		if n.Name != "root" {
			valueStr := ""
			if n.Value != nil {
				valueStr = fmt.Sprintf("%v", n.Value)
			}
			if n.IsRepeated {
				valueStr += " [repeated]"
			}

			rows = append(rows, TableRow{
				FieldName: n.Name,
				FieldNum:  n.FieldNum,
				Type:      n.Type,
				Value:     valueStr,
				Children:  len(n.Children),
			})
		}

		// Рекурсивно обрабатываем дочерние узлы
		for _, child := range n.Children {
			traverse(child, level+1)
		}
	}

	traverse(node, 0)
	return rows
}

// createTableWidget создает виджет таблицы из данных
func createTableWidget(data []TableRow) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("(no data)")
	}

	// Создаем заголовки таблицы (жирным шрифтом)
	headerName := widget.NewLabel("Field Name")
	headerName.TextStyle = fyne.TextStyle{Bold: true}
	headerNum := widget.NewLabel("Field #")
	headerNum.TextStyle = fyne.TextStyle{Bold: true}
	headerType := widget.NewLabel("Type")
	headerType.TextStyle = fyne.TextStyle{Bold: true}
	headerValue := widget.NewLabel("Value")
	headerValue.TextStyle = fyne.TextStyle{Bold: true}
	headerChildren := widget.NewLabel("Children")
	headerChildren.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewGridWithColumns(5,
		headerName,
		headerNum,
		headerType,
		headerValue,
		headerChildren,
	)

	// Создаем строки таблицы
	rows := make([]fyne.CanvasObject, 0, len(data))
	for _, row := range data {
		rowContainer := container.NewGridWithColumns(5,
			widget.NewLabel(row.FieldName),
			widget.NewLabel(fmt.Sprintf("%d", row.FieldNum)),
			widget.NewLabel(row.Type),
			widget.NewLabel(row.Value),
			widget.NewLabel(fmt.Sprintf("%d", row.Children)),
		)
		rows = append(rows, rowContainer)
	}

	// Объединяем заголовок и строки
	content := container.NewVBox(header)
	for _, row := range rows {
		content.Add(row)
	}

	return content
}
