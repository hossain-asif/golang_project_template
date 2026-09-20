package main

import (
	"log"

	"go_project_structure/app"
	"go_project_structure/config/env"
)

func main() {
	env.Load()

	cfg := app.NewConfig()
	application := app.NewApplication(cfg)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
