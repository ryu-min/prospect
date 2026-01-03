package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Проверяем наличие protoc
	protocPath, err := findProtoc()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: protoc не найден: %v\n", err)
		fmt.Fprintln(os.Stderr, "Установите protoc: scoop install protobuf")
		os.Exit(1)
	}

	testdataDir := "testdata"
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания директории: %v\n", err)
		os.Exit(1)
	}

	// Создаем простой бинарный файл для тестирования
	// Используем protoc для кодирования тестовых данных

	// Создаем простой текстовый protobuf файл для кодирования
	simpleText := `text: "Hello, World!"
number: 42
flag: true
`

	simpleTextFile := filepath.Join(testdataDir, "simple_text.txt")
	if err := os.WriteFile(simpleTextFile, []byte(simpleText), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи файла: %v\n", err)
		os.Exit(1)
	}

	// Кодируем через protoc
	simpleProto := filepath.Join(testdataDir, "simple.proto")
	simpleBinary := filepath.Join(testdataDir, "simple.bin")

	cmd := exec.Command(protocPath,
		"--encode", "test.SimpleMessage",
		"--proto_path", testdataDir,
		simpleProto,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Для создания бинарного файла выполните:\n")
	fmt.Printf("  cat %s | %s --encode test.SimpleMessage --proto_path %s %s > %s\n",
		simpleTextFile, protocPath, testdataDir, simpleProto, simpleBinary)

	// Создаем более сложный пример
	personText := `name: "John Doe"
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
`

	personTextFile := filepath.Join(testdataDir, "person_text.txt")
	if err := os.WriteFile(personTextFile, []byte(personText), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи файла: %v\n", err)
		os.Exit(1)
	}

	personProto := filepath.Join(testdataDir, "person.proto")
	personBinary := filepath.Join(testdataDir, "person.bin")

	fmt.Printf("\nДля создания бинарного файла person выполните:\n")
	fmt.Printf("  cat %s | %s --encode test.Person --proto_path %s %s > %s\n",
		personTextFile, protocPath, testdataDir, personProto, personBinary)

	fmt.Println("\nИли создайте бинарные файлы вручную используя protoc --encode")
}

func findProtoc() (string, error) {
	cmd := exec.Command("protoc", "--version")
	if err := cmd.Run(); err == nil {
		return "protoc", nil
	}
	return "", fmt.Errorf("protoc не найден")
}

