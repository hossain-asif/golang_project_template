package main

import (
	"log"

	"go_project_structure/app"
	config "go_project_structure/config/env"
	"go_project_structure/internal/router"
)

func main() {
	config.Load()

	cfg := app.NewConfig()
	application := app.NewApplication(cfg, router.Modules)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
