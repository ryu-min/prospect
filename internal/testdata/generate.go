package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	testdataDir := "."
	protocPath := "protoc"

	// Проверяем protoc
	cmd := exec.Command(protocPath, "--version")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: protoc не найден. Установите через: scoop install protobuf\n")
		os.Exit(1)
	}

	fmt.Println("Создание бинарных файлов...")

	// Создаем simple.bin
	simpleText := `text: "Hello, World!"
number: 42
flag: true
`
	simpleProto := filepath.Join(testdataDir, "simple.proto")
	simpleBinary := filepath.Join(testdataDir, "simple.bin")

	encodeCmd := exec.Command(protocPath, "--encode", "test.SimpleMessage", "--proto_path", testdataDir, simpleProto)
	encodeCmd.Stdin = bytes.NewReader([]byte(simpleText))
	
	output, err := encodeCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка кодирования simple: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(simpleBinary, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи simple.bin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Создан simple.bin (%d байт)\n", len(output))

	// Создаем person.bin
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

	personProto := filepath.Join(testdataDir, "person.proto")
	personBinary := filepath.Join(testdataDir, "person.bin")

	encodeCmd = exec.Command(protocPath, "--encode", "test.Person", "--proto_path", testdataDir, personProto)
	encodeCmd.Stdin = bytes.NewReader([]byte(personText))
	
	output, err = encodeCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка кодирования person: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(personBinary, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи person.bin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Создан person.bin (%d байт)\n", len(output))
	fmt.Println("Готово!")
}
