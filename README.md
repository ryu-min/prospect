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

**Доступные функции:**
- `ui.CreateTab(tabs)` - создает новую вкладку по умолчанию
- `ui.AddTab(tabs, name, content)` - универсальная функция добавления вкладки с произвольным содержимым

## Использование

### Запуск приложения

```bash
# Напрямую
$env:CGO_ENABLED=1; go run ./cmd/prospect
```

### Создание вкладки с произвольным содержимым

Используйте функцию `AddTab` для создания вкладки с любым содержимым:

```go
import "prospect/internal/ui"

tabs := container.NewAppTabs()

// Создание вкладки с произвольным содержимым
content := widget.NewLabel("Мое содержимое")
ui.AddTab(tabs, "Моя вкладка", content)
```

### Работа с вкладками напрямую

```go
import "prospect/internal/ui"

// Создание контейнера вкладок
tabs := container.NewAppTabs()

// Добавление вкладки с произвольным содержимым
ui.AddTab(tabs, "Имя вкладки", widget.NewLabel("Содержимое"))

// Создание вкладки по умолчанию
ui.CreateTab(tabs)
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
