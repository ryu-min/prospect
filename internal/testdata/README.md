# Тестовые данные для Prospect

Эта папка содержит тестовые protobuf файлы и бинарные данные для тестирования приложения.

## Файлы

- `simple.proto` - Простая proto схема с тремя полями (text, number, flag)
- `person.proto` - Более сложная схема с вложенными сообщениями
- `simple.bin` - Бинарный файл, закодированный из simple.proto
- `person.bin` - Бинарный файл, закодированный из person.proto

## Создание тестовых данных

Для создания бинарных файлов выполните:

```bash
cd testdata
go run generate.go
```

Это создаст `simple.bin` и `person.bin` из соответствующих proto схем.

## Использование в приложении

1. Запустите приложение Prospect
2. Откройте вкладку "Protobuf Viewer"
3. Нажмите "Открыть бинарный protobuf файл"
4. Выберите один из `.bin` файлов из этой папки
5. Приложение декодирует файл через `protoc --decode_raw` и отобразит дерево

## Применение схемы

После открытия бинарного файла вы можете применить proto схему:

1. Нажмите "Применить proto схему"
2. Выберите соответствующий `.proto` файл
3. Приложение попытается декодировать данные с использованием схемы

## Формат данных

### simple.bin
Содержит:
- field_1: "Hello, World!" (string)
- field_2: 42 (number)
- field_3: 1 (bool)

### person.bin
Содержит:
- field_1: "John Doe" (string)
- field_2: 30 (number)
- field_3: "john@example.com" (string)
- field_4: message (address)
  - field_1: "123 Main St" (string)
  - field_2: "New York" (string)
  - field_3: "USA" (string)
  - field_4: 10001 (number)
- field_5: ["reading", "coding"] (repeated string)

