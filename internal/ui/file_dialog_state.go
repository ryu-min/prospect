package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type fileDialogState struct {
	lastDirPath string
	dialogSize  fyne.Size
}

var globalDialogState *fileDialogState

func getFileDialogState() *fileDialogState {
	if globalDialogState == nil {
		loadFileDialogState()
		if globalDialogState == nil {
			wd, _ := os.Getwd()
			globalDialogState = &fileDialogState{
				dialogSize:  fyne.NewSize(800, 600),
				lastDirPath: wd,
			}
		}
	}
	return globalDialogState
}

func (fds *fileDialogState) setLastDir(uri fyne.URI) {
	if uri == nil {
		return
	}

	path := uri.Path()
	dir := filepath.Dir(path)

	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastDirPath = dir
		saveFileDialogState()
	}
}

func (fds *fileDialogState) setLastOpenDir(uri fyne.URI) {
	fds.setLastDir(uri)
}

func (fds *fileDialogState) setLastSaveDir(uri fyne.URI) {
	fds.setLastDir(uri)
}

func (fds *fileDialogState) setLastSchemaDir(uri fyne.URI) {
	// Используем общий lastDirPath для всех диалогов
	fds.setLastDir(uri)
}

func (fds *fileDialogState) getLastDir() fyne.ListableURI {
	dirPath := fds.lastDirPath
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

func (fds *fileDialogState) getLastOpenDir() fyne.ListableURI {
	return fds.getLastDir()
}

func (fds *fileDialogState) getLastSaveDir() fyne.ListableURI {
	return fds.getLastDir()
}

func (fds *fileDialogState) getLastSchemaDir() fyne.ListableURI {
	// Используем общий lastDirPath для всех диалогов
	return fds.getLastDir()
}

func (fds *fileDialogState) setDialogSize(size fyne.Size) {
	fds.dialogSize = size
	saveFileDialogState()
}

func (fds *fileDialogState) getDialogSize() fyne.Size {
	return fds.dialogSize
}
