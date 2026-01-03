package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	data, err := os.ReadFile("simple.bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("protoc", "--decode_raw")
	cmd.Stdin = bytes.NewReader(data)
	
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка декодирования: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Вывод protoc --decode_raw:")
	fmt.Println(string(output))
	fmt.Println("\n--- Hex dump ---")
	for i, b := range data {
		if i%16 == 0 {
			fmt.Printf("\n%04x: ", i)
		}
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
}
