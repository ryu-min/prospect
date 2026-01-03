package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// FileDialogState управляет состоянием диалогов файлов
type FileDialogState struct {
	lastOpenDirPath   string
	lastSaveDirPath   string
	lastSchemaDirPath string
	dialogSize        fyne.Size
}

var globalDialogState *FileDialogState

// GetFileDialogState возвращает глобальное состояние диалогов
func GetFileDialogState() *FileDialogState {
	if globalDialogState == nil {
		// Инициализируем с текущей рабочей директорией
		wd, _ := os.Getwd()
		globalDialogState = &FileDialogState{
			dialogSize:        fyne.NewSize(800, 600), // Размер по умолчанию
			lastOpenDirPath:   wd,
			lastSaveDirPath:   wd,
			lastSchemaDirPath: wd,
		}
	}
	return globalDialogState
}

// SetLastOpenDir сохраняет последнюю директорию для открытия файлов
func (fds *FileDialogState) SetLastOpenDir(uri fyne.URI) {
	if uri == nil {
		return
	}
	
	// Получаем директорию из URI
	path := uri.Path()
	dir := filepath.Dir(path)
	
	// Проверяем, что директория существует
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastOpenDirPath = dir
	}
}

// SetLastSaveDir сохраняет последнюю директорию для сохранения файлов
func (fds *FileDialogState) SetLastSaveDir(uri fyne.URI) {
	if uri == nil {
		return
	}
	
	path := uri.Path()
	dir := filepath.Dir(path)
	
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastSaveDirPath = dir
	}
}

// SetLastSchemaDir сохраняет последнюю директорию для схем
func (fds *FileDialogState) SetLastSchemaDir(uri fyne.URI) {
	if uri == nil {
		return
	}
	
	path := uri.Path()
	dir := filepath.Dir(path)
	
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		fds.lastSchemaDirPath = dir
	}
}

// GetLastOpenDir возвращает последнюю директорию для открытия
func (fds *FileDialogState) GetLastOpenDir() fyne.ListableURI {
	dirPath := fds.lastOpenDirPath
	if dirPath == "" {
		wd, _ := os.Getwd()
		dirPath = wd
	}
	
	// Проверяем, что директория существует
	if info, err := os.Stat(dirPath); err != nil || !info.IsDir() {
		wd, _ := os.Getwd()
		dirPath = wd
	}
	
	uri := storage.NewFileURI(dirPath)
	if listableURI, err := storage.ListerForURI(uri); err == nil {
		return listableURI
	}
	
	// Fallback - возвращаем URI текущей директории
	wd, _ := os.Getwd()
	if listableURI, err := storage.ListerForURI(storage.NewFileURI(wd)); err == nil {
		return listableURI
	}
	return nil
}

// GetLastSaveDir возвращает последнюю директорию для сохранения
func (fds *FileDialogState) GetLastSaveDir() fyne.ListableURI {
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

// GetLastSchemaDir возвращает последнюю директорию для схем
func (fds *FileDialogState) GetLastSchemaDir() fyne.ListableURI {
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

// SetDialogSize устанавливает размер диалога
func (fds *FileDialogState) SetDialogSize(size fyne.Size) {
	fds.dialogSize = size
}

// GetDialogSize возвращает размер диалога
func (fds *FileDialogState) GetDialogSize() fyne.Size {
	return fds.dialogSize
}

