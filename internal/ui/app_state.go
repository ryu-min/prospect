package ui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

type appState struct {
	Tabs            []tabState `json:"tabs"`
	SelectedTab     int        `json:"selectedTab"`
	FileDialogState struct {
		LastDirPath  string  `json:"lastDirPath"`
		DialogWidth  float32 `json:"dialogWidth"`
		DialogHeight float32 `json:"dialogHeight"`
	} `json:"fileDialogState"`
}

type tabState struct {
	Title             string `json:"title"`
	FilePath          string `json:"filePath,omitempty"`
	SchemaPath        string `json:"schemaPath,omitempty"`
	SchemaMessageName string `json:"schemaMessageName,omitempty"`
}

func getAppStatePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "prospect")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "state.json"), nil
}

func loadAppState() (*appState, error) {
	statePath, err := getAppStatePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &appState{
				Tabs:        make([]tabState, 0),
				SelectedTab: -1,
			}, nil
		}
		return nil, err
	}

	var state appState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func saveAppState(state *appState) error {
	statePath, err := getAppStatePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

func saveTabState(tm *tabManager) error {
	state := &appState{
		Tabs:        make([]tabState, 0, len(tm.tabs)),
		SelectedTab: tm.selectedTab,
	}

	for _, tab := range tm.tabs {
		state.Tabs = append(state.Tabs, tabState{
			Title:             tab.title,
			FilePath:          tab.filePath,
			SchemaPath:        tab.schemaPath,
			SchemaMessageName: tab.schemaMessageName,
		})
	}

	fileDialogState := getFileDialogState()
	state.FileDialogState.LastDirPath = fileDialogState.lastDirPath
	state.FileDialogState.DialogWidth = fileDialogState.dialogSize.Width
	state.FileDialogState.DialogHeight = fileDialogState.dialogSize.Height

	return saveAppState(state)
}

func loadTabState(tm *tabManager, fyneApp fyne.App, window fyne.Window) error {
	state, err := loadAppState()
	if err != nil {
		return err
	}

	if len(state.Tabs) == 0 {
		return nil
	}

	fileDialogState := getFileDialogState()
	if state.FileDialogState.LastDirPath != "" {
		fileDialogState.lastDirPath = state.FileDialogState.LastDirPath
	}
	if state.FileDialogState.DialogWidth > 0 && state.FileDialogState.DialogHeight > 0 {
		fileDialogState.dialogSize = fyne.NewSize(state.FileDialogState.DialogWidth, state.FileDialogState.DialogHeight)
	}

	for i, tabState := range state.Tabs {
		content := protoViewWithFile(fyneApp, window, tm, tabState.FilePath, tabState.SchemaPath, tabState.SchemaMessageName)
		tabIndex := len(tm.tabs)
		tm.addTabWithPathWithoutSave(tabState.Title, content, tabState.FilePath)
		if tabState.SchemaPath != "" {
			tm.tabs[tabIndex].schemaPath = tabState.SchemaPath
			tm.tabs[tabIndex].schemaMessageName = tabState.SchemaMessageName
		}
		if i == state.SelectedTab && state.SelectedTab >= 0 && state.SelectedTab < len(state.Tabs) {
			tm.selectTabWithoutSave(tabIndex)
		}
	}

	return nil
}

func saveFileDialogState() {
	if globalDialogState == nil {
		return
	}
	state, _ := loadAppState()
	if state == nil {
		state = &appState{
			Tabs:        make([]tabState, 0),
			SelectedTab: -1,
		}
	}
	state.FileDialogState.LastDirPath = globalDialogState.lastDirPath
	state.FileDialogState.DialogWidth = globalDialogState.dialogSize.Width
	state.FileDialogState.DialogHeight = globalDialogState.dialogSize.Height
	saveAppState(state)
}

func loadFileDialogState() {
	state, err := loadAppState()
	if err != nil || state == nil {
		return
	}

	wd, _ := os.Getwd()
	globalDialogState = &fileDialogState{
		dialogSize:  fyne.NewSize(800, 600),
		lastDirPath: wd,
	}

	if state.FileDialogState.LastDirPath != "" {
		globalDialogState.lastDirPath = state.FileDialogState.LastDirPath
	}
	if state.FileDialogState.DialogWidth > 0 && state.FileDialogState.DialogHeight > 0 {
		globalDialogState.dialogSize = fyne.NewSize(state.FileDialogState.DialogWidth, state.FileDialogState.DialogHeight)
	}
}
