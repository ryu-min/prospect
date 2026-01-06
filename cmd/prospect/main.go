package main

import (
	"fmt"
	"os"

	"prospect/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Ошибка запуска приложения: %v\n", err)
		os.Exit(1)
	}
}
