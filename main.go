package main

import (
	"go_project_structure/app"
	config "go_project_structure/config/env"
)

func main() {
	config.Load()
	cfg := app.NewConfig()
	app := app.NewApplication(cfg)

	app.Run()
}
