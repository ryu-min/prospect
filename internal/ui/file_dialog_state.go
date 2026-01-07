package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type fileDialogState struct {
	lastOpenDirPath   string
	lastSaveDirPath   string
	lastSchemaDirPath string
	dialogSize        fyne.Size
}

var globalDialogState *fileDialogState

func getFileDialogState() *fileDialogState {
	if globalDialogState == nil {
		wd, _ := os.Getwd()
		globalDialogState = &fileDialogState{
			dialogSize:        fyne.NewSize(800, 600),
			lastOpenDirPath:   wd,
			lastSaveDirPath:   wd,
			lastSchemaDirPath: wd,
		}
	}
	return globalDialogState
}

func (fds *fileDialogState) setLastOpenDir(uri fyne.URI) {
	if uri == nil {
		return
	}

	path := uri.Path()
	dir := filepath.Dir(path)

	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastOpenDirPath = dir
	}
}

func (fds *fileDialogState) setLastSaveDir(uri fyne.URI) {
	if uri == nil {
		return
	}

	path := uri.Path()
	dir := filepath.Dir(path)

	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastSaveDirPath = dir
	}
}

func (fds *fileDialogState) setLastSchemaDir(uri fyne.URI) {
	if uri == nil {
		return
	}

	path := uri.Path()
	dir := filepath.Dir(path)

	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastSchemaDirPath = dir
	}
}

func (fds *fileDialogState) getLastOpenDir() fyne.ListableURI {
	dirPath := fds.lastOpenDirPath
	if dirPath == "" {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	uri := storage.NewFileURI(dirPath)
	if listableURI, err := storage.ListerForURI(uri); err == nil {
		return listableURI
	}

	wd, _ := os.Getwd()
	if listableURI, err := storage.ListerForURI(storage.NewFileURI(wd)); err == nil {
		return listableURI
	}
	return nil
}

func (fds *fileDialogState) getLastSaveDir() fyne.ListableURI {
	dirPath := fds.lastSaveDirPath
	if dirPath == "" {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	uri := storage.NewFileURI(dirPath)
	if listableURI, err := storage.ListerForURI(uri); err == nil {
		return listableURI
	}

	wd, _ := os.Getwd()
	if listableURI, err := storage.ListerForURI(storage.NewFileURI(wd)); err == nil {
		return listableURI
	}
	return nil
}

func (fds *fileDialogState) getLastSchemaDir() fyne.ListableURI {
	dirPath := fds.lastSchemaDirPath
	if dirPath == "" {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		wd, _ := os.Getwd()
		dirPath = wd
	}

	uri := storage.NewFileURI(dirPath)
	if listableURI, err := storage.ListerForURI(uri); err == nil {
		return listableURI
	}

	wd, _ := os.Getwd()
	if listableURI, err := storage.ListerForURI(storage.NewFileURI(wd)); err == nil {
		return listableURI
	}
	return nil
}

func (fds *fileDialogState) setDialogSize(size fyne.Size) {
	fds.dialogSize = size
}

func (fds *fileDialogState) getDialogSize() fyne.Size {
	return fds.dialogSize
}
