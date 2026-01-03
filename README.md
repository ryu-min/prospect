# Prospect - Система управления вкладками

Приложение на Go с использованием Fyne для управления вкладками с различными виджетами.

## Структура проекта

```
prospect/
├── cmd/
│   └── prospect/
│       └── main.go          # Точка входа приложения
├── internal/
│   ├── app/
│   │   └── app.go           # Основной класс приложения
│   └── ui/
│       └── window.go        # Создание окна и функции управления вкладками
├── go.mod
├── go.sum
└── README.md
```

## Архитектура

### Управление вкладками

Приложение использует `container.AppTabs` напрямую из Fyne для управления вкладками. Все функции создания вкладок находятся в пакете `internal/ui`.

**Доступные функции создания вкладок:**
- `ui.CreateTextTab(tabs)` - вкладка с текстом
- `ui.CreateFormTab(tabs)` - вкладка с формой
- `ui.CreateListTab(tabs)` - вкладка со списком
- `ui.CreateInputTab(tabs)` - вкладка с элементами ввода
- `ui.CreateProgressTab(tabs)` - вкладка с прогресс-барами
- `ui.CreateCustomTab(tabs)` - кастомная вкладка
- `ui.AddTab(tabs, name, content)` - универсальная функция добавления вкладки

## Использование

### Запуск приложения

```bash
# Напрямую
$env:CGO_ENABLED=1; go run ./cmd/prospect
```

### Добавление нового типа вкладки

1. Создайте функцию в `internal/ui/window.go`:

```go
func CreateMyTab(tabs *container.AppTabs) {
    tabCounter++
    tabName := fmt.Sprintf("Моя вкладка #%d", tabCounter)
    
    content := widget.NewLabel("Содержимое вкладки")
    AddTab(tabs, tabName, content)
}
```

2. Добавьте тип в список в функциях `createControlTab` и `createToolbar`:

```go
tabTypes := []struct {
    name string
    fn   func(*container.AppTabs)
}{
    // ... существующие типы
    {"Моя вкладка", CreateMyTab},
}

typeMap := map[string]func(*container.AppTabs){
    // ... существующие маппинги
    "Моя вкладка": CreateMyTab,
}
```

### Работа с вкладками напрямую

```go
import "prospect/internal/ui"

// Создание контейнера вкладок
tabs := container.NewAppTabs()

// Добавление вкладки с произвольным содержимым
ui.AddTab(tabs, "Имя вкладки", widget.NewLabel("Содержимое"))

// Создание вкладки по типу
ui.CreateTextTab(tabs)
```

## Требования

- Go 1.21+
- Компилятор C (gcc) для Windows (для работы Fyne)
- Fyne v2.4.5+

## Установка зависимостей

```bash
go mod tidy
```

## Сборка

```bash
$env:CGO_ENABLED=1; go build -o prospect.exe ./cmd/prospect
```
