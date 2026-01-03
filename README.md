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
│   ├── ui/
│   │   └── window.go        # Создание и настройка главного окна
│   └── widgets/
│       ├── tabmanager.go    # Менеджер вкладок (универсальный)
│       └── tabtypes.go      # Типы вкладок (фабрики)
├── go.mod
├── go.sum
└── README.md
```

## Архитектура

### TabManager (internal/widgets/tabmanager.go)

Базовый виджет-менеджер для управления вкладками. Работает с любыми виджетами через интерфейс `TabCreator`.

**Основные возможности:**
- Добавление вкладок с произвольным содержимым
- Регистрация создателей вкладок по типу
- Создание вкладок по зарегистрированному типу
- Удаление текущей вкладки

**Интерфейс TabCreator:**
```go
type TabCreator interface {
    CreateContent() fyne.CanvasObject
    GetName() string
}
```

### Типы вкладок (internal/widgets/tabtypes.go)

Реализации `TabCreator` для различных типов вкладок:
- `TextTabCreator` - вкладка с текстом
- `FormTabCreator` - вкладка с формой
- `ListTabCreator` - вкладка со списком
- `InputTabCreator` - вкладка с элементами ввода
- `ProgressTabCreator` - вкладка с прогресс-барами
- `CustomTabCreator` - кастомная вкладка

## Использование

### Запуск приложения

```bash
# Через скрипт
.\run.ps1    # PowerShell
.\run.bat    # CMD

# Или напрямую
$env:CGO_ENABLED=1; go run ./cmd/prospect
```

### Добавление нового типа вкладки

1. Создайте структуру, реализующую интерфейс `TabCreator`:

```go
type MyTabCreator struct {
    BaseTabCreator
}

func NewMyTabCreator() *MyTabCreator {
    return &MyTabCreator{
        BaseTabCreator: BaseTabCreator{name: "Моя вкладка"},
    }
}

func (mtc *MyTabCreator) CreateContent() fyne.CanvasObject {
    // Создайте и верните содержимое вкладки
    return widget.NewLabel("Содержимое")
}
```

2. Зарегистрируйте создатель в `internal/ui/window.go`:

```go
tabManager.RegisterTabCreator("Моя вкладка", widgets.NewMyTabCreator())
```

### Работа с TabManager напрямую

```go
// Создание менеджера
tabManager := widgets.NewTabManager()

// Добавление вкладки с произвольным содержимым
tabManager.AddTab("Имя вкладки", widget.NewLabel("Содержимое"))

// Регистрация создателя
tabManager.RegisterTabCreator("Тип", creator)

// Создание вкладки по типу
tabManager.CreateTabByType("Тип")
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

