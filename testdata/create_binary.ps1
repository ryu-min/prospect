# Скрипт для создания бинарных protobuf файлов

$testdataDir = "."
$protocPath = "protoc"

# Проверяем наличие protoc
$protocVersion = & $protocPath --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка: protoc не найден. Установите через: scoop install protobuf"
    exit 1
}

Write-Host "Создание бинарных файлов..."

# Создаем simple.bin
$simpleText = @"
text: "Hello, World!"
number: 42
flag: true
"@

$simpleText | & $protocPath --encode test.SimpleMessage --proto_path $testdataDir simple.proto | Out-File -FilePath "simple.bin" -Encoding binary
Write-Host "Создан simple.bin"

# Создаем person.bin
$personText = @"
name: "John Doe"
age: 30
email: "john@example.com"
address {
  street: "123 Main St"
  city: "New York"
  country: "USA"
  zip_code: 10001
}
hobbies: "reading"
hobbies: "coding"
"@

$personText | & $protocPath --encode test.Person --proto_path $testdataDir person.proto | Out-File -FilePath "person.bin" -Encoding binary
Write-Host "Создан person.bin"

Write-Host "Готово!"

