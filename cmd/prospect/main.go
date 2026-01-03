package main

import (
	"fmt"
	"os"

	"prospect/internal/app"
)

func main() {
	fmt.Fprintln(os.Stdout, "[INFO] Запуск приложения Prospect...")

	// Создаем и запускаем приложение
	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Ошибка запуска приложения: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "[INFO] Приложение завершено")
}

