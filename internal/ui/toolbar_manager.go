package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type toolbarManager struct {
	toolbar        fyne.CanvasObject
	openBtn        *widget.Button
	saveBtn        *widget.Button
	applySchemaBtn *widget.Button
	exportJSONBtn  *widget.Button
}

func newToolbarManager() *toolbarManager {
	tm := &toolbarManager{}

	tm.openBtn = widget.NewButtonWithIcon("Open binary", theme.FolderOpenIcon(), func() {})
	tm.openBtn.Importance = widget.LowImportance

	tm.saveBtn = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {})
	tm.saveBtn.Importance = widget.LowImportance

	tm.applySchemaBtn = widget.NewButtonWithIcon("Apply schema", theme.SettingsIcon(), func() {})
	tm.applySchemaBtn.Importance = widget.LowImportance

	tm.exportJSONBtn = widget.NewButtonWithIcon("Export schema", theme.FileIcon(), func() {})
	tm.exportJSONBtn.Importance = widget.LowImportance

	tm.toolbar = container.NewHBox(
		tm.openBtn,
		tm.saveBtn,
		tm.applySchemaBtn,
		tm.exportJSONBtn,
	)
	return tm
}

func (tm *toolbarManager) SetOpenCallback(callback func()) {
	tm.openBtn.OnTapped = callback
}

func (tm *toolbarManager) SetSaveCallback(callback func()) {
	tm.saveBtn.OnTapped = callback
}

func (tm *toolbarManager) SetApplySchemaCallback(callback func()) {
	tm.applySchemaBtn.OnTapped = callback
}

func (tm *toolbarManager) SetExportJSONCallback(callback func()) {
	tm.exportJSONBtn.OnTapped = callback
}

func (tm *toolbarManager) GetToolbar() fyne.CanvasObject {
	return tm.toolbar
}
