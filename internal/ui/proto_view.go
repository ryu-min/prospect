package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"prospect/internal/protobuf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func protoView(fyneApp fyne.App, parentWindow fyne.Window, browserTabs *tabManager) fyne.CanvasObject {
	return protoViewWithFile(fyneApp, parentWindow, browserTabs, "")
}

func protoViewWithFile(fyneApp fyne.App, parentWindow fyne.Window, browserTabs *tabManager, filePath string) fyne.CanvasObject {
	parser, err := protobuf.NewParser()
	if err != nil {
		errorLabel := widget.NewLabel(fmt.Sprintf("Error: %v", err))
		errorLabel.Wrapping = fyne.TextWrapWord
		return container.NewPadded(errorLabel)
	}

	var currentTree *protobuf.TreeNode
	var currentFilePath string
	if filePath != "" {
		currentFilePath = filePath
	}

	treeWidget := createProtoTree(nil)

	treeScrollContainer := container.NewScroll(treeWidget)

	dialogState := getFileDialogState()

	var openBtn *widget.Button
	var saveBtn *widget.Button
	var applySchemaBtn *widget.Button
	var exportJSONBtn *widget.Button
	var buttonPanel fyne.CanvasObject

	openBtn = widget.NewButton("Open binary proto file", func() {
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

			currentFilePath = reader.URI().Path()

			if browserTabs != nil {
				fileName := filepath.Base(currentFilePath)
				browserTabs.UpdateTabTitle(fileName)
				browserTabs.SetTabFilePath(currentFilePath)
			}

			dialogState.setLastOpenDir(reader.URI())

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

			log.Printf("Parsing proto file: %s", reader.URI().Path())

			tree, err := parser.ParseRaw(data)
			if err != nil {
				dialog.ShowError(fmt.Errorf("parsing error: %w", err), parentWindow)
				return
			}

			currentTree = tree

			adapter := newProtoTreeAdapter(tree)
			adapter.SetWindow(parentWindow)

			newTreeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
			adapter.SetTreeWidget(newTreeWidget)

			rootChildren := adapter.ChildUIDs("")

			if len(rootChildren) > 0 {
				newTreeWidget.OpenBranch("")
			}

			newTreeWidget.Refresh()

			treeWidget = newTreeWidget

			newScrollContainer := container.NewScroll(newTreeWidget)
			newScrollContainer.Refresh()
			treeScrollContainer = newScrollContainer

			_ = adapter.ChildUIDs("root")
			_ = adapter.IsBranch("root")

			newBorder := container.NewBorder(
				buttonPanel,
				nil,
				nil,
				nil,
				newScrollContainer,
			)
			if browserTabs != nil {
				browserTabs.UpdateTabContent(container.NewPadded(newBorder))
			} else {
				log.Printf("Error: browserTabs is nil")
			}
			log.Printf("Proto file parsed successfully, tree updated")
		}, parentWindow)

		if lastDir := dialogState.getLastOpenDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		fileDialog.Resize(dialogState.getDialogSize())
		fileDialog.Show()
	})

	applySchemaBtn = widget.NewButton("Apply proto schema", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a proto file first", parentWindow)
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

			dialogState.setLastSchemaDir(reader.URI())

			schemaPath := reader.URI().Path()
			log.Printf("Applying schema: %s", schemaPath)

			tree, err := parser.ApplySchema(currentTree, schemaPath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("error applying schema: %w", err), parentWindow)
				return
			}

			currentTree = tree
			adapter := newProtoTreeAdapter(tree)
			adapter.SetWindow(parentWindow) // Устанавливаем окно для диалогов
			newTreeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
			newTreeWidget.OpenBranch("root")
			treeWidget = newTreeWidget
			newScrollContainer := container.NewScroll(newTreeWidget)
			treeScrollContainer = newScrollContainer
			newBorder := container.NewBorder(
				buttonPanel,
				nil,
				nil,
				nil,
				newScrollContainer,
			)
			if browserTabs != nil {
				browserTabs.UpdateTabContent(container.NewPadded(newBorder))
			}
			log.Printf("Schema applied successfully, tree updated")
		}, parentWindow)

		if lastDir := dialogState.getLastSchemaDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		fileDialog.Resize(dialogState.getDialogSize())
		fileDialog.Show()
	})

	saveBtn = widget.NewButton("Save", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a proto file first", parentWindow)
			return
		}

		serializer := protobuf.NewSerializer(parser.GetProtocPath())
		binaryData, err := serializer.SerializeRaw(currentTree)
		if err != nil {
			dialog.ShowError(fmt.Errorf("serialization error: %w", err), parentWindow)
			return
		}

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

			currentFilePath = writer.URI().Path()

			if browserTabs != nil {
				browserTabs.SetTabFilePath(currentFilePath)
				fileName := filepath.Base(currentFilePath)
				browserTabs.UpdateTabTitle(fileName)
			}

			dialogState.setLastSaveDir(writer.URI())

			if _, err := writer.Write(binaryData); err != nil {
				dialog.ShowError(fmt.Errorf("write error: %w", err), parentWindow)
				return
			}

			dialog.ShowInformation("Success", "Proto file saved", parentWindow)
			log.Printf("Proto file saved: %s", currentFilePath)
		}, parentWindow)

		if lastDir := dialogState.getLastSaveDir(); lastDir != nil {
			saveDialog.SetLocation(lastDir)
		} else if currentFilePath != "" {
			dirPath := filepath.Dir(currentFilePath)
			uri := storage.NewFileURI(dirPath)
			if listableURI, err := storage.ListerForURI(uri); err == nil {
				saveDialog.SetLocation(listableURI)
			}
		}

		saveDialog.Resize(dialogState.getDialogSize())
		saveDialog.Show()
	})

	exportJSONBtn = widget.NewButton("Export to JSON", func() {
		if currentTree == nil {
			dialog.ShowInformation("Information", "Please open a proto file first", parentWindow)
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

			dialogState.setLastSaveDir(writer.URI())

			if _, err := writer.Write(jsonData); err != nil {
				dialog.ShowError(fmt.Errorf("write error: %w", err), parentWindow)
				return
			}

			dialog.ShowInformation("Success", "JSON file saved", parentWindow)
		}, parentWindow)

		if lastDir := dialogState.getLastSaveDir(); lastDir != nil {
			fileDialog.SetLocation(lastDir)
		}

		fileDialog.Resize(dialogState.getDialogSize())
		fileDialog.Show()
	})

	buttonPanel = container.NewHBox(
		openBtn,
		saveBtn,
		applySchemaBtn,
		exportJSONBtn,
	)

	if filePath != "" {
		if err := loadFileIntoView(filePath, parser, &currentTree, &treeWidget, &treeScrollContainer, parentWindow, browserTabs, dialogState, &currentFilePath); err != nil {
			log.Printf("Failed to load file %s: %v", filePath, err)
		}
	}

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

func loadFileIntoView(filePath string, parser *protobuf.Parser, currentTree **protobuf.TreeNode, treeWidget **widget.Tree, treeScrollContainer **container.Scroll, parentWindow fyne.Window, browserTabs *tabManager, dialogState *fileDialogState, currentFilePath *string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	*currentFilePath = filePath

	if browserTabs != nil {
		browserTabs.SetTabFilePath(filePath)
	}

	dialogState.setLastOpenDir(storage.NewFileURI(filePath))

	log.Printf("Parsing proto file: %s", filePath)

	tree, err := parser.ParseRaw(data)
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}

	*currentTree = tree

	adapter := newProtoTreeAdapter(tree)
	adapter.SetWindow(parentWindow)

	newTreeWidget := widget.NewTree(adapter.ChildUIDs, adapter.IsBranch, adapter.CreateNode, adapter.UpdateNode)
	adapter.SetTreeWidget(newTreeWidget)

	rootChildren := adapter.ChildUIDs("")

	if len(rootChildren) > 0 {
		newTreeWidget.OpenBranch("")
	}

	newTreeWidget.Refresh()

	*treeWidget = newTreeWidget

	newScrollContainer := container.NewScroll(newTreeWidget)
	newScrollContainer.Refresh()
	*treeScrollContainer = newScrollContainer

	return nil
}

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
		result += "Proto Root\n"
		result += strings.Repeat("=", 50) + "\n"
		if len(node.Children) == 0 {
			result += "(no data)\n"
			return result
		}
	} else {
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

	for _, child := range node.Children {
		result += formatTree(child, indent+1)
	}

	return result
}

type tableRow struct {
	FieldName string
	FieldNum  int
	Type      string
	Value     string
	Children  int
}

func buildTableData(node *protobuf.TreeNode) []tableRow {
	var rows []tableRow

	var traverse func(*protobuf.TreeNode, int)
	traverse = func(n *protobuf.TreeNode, level int) {
		if n == nil {
			return
		}

		if n.Name != "root" {
			valueStr := ""
			if n.Value != nil {
				valueStr = fmt.Sprintf("%v", n.Value)
			}
			if n.IsRepeated {
				valueStr += " [repeated]"
			}

			rows = append(rows, tableRow{
				FieldName: n.Name,
				FieldNum:  n.FieldNum,
				Type:      n.Type,
				Value:     valueStr,
				Children:  len(n.Children),
			})
		}

		for _, child := range n.Children {
			traverse(child, level+1)
		}
	}

	traverse(node, 0)
	return rows
}

func createTableWidget(data []tableRow) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("(no data)")
	}

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

	content := container.NewVBox(header)
	for _, row := range rows {
		content.Add(row)
	}

	return content
}
