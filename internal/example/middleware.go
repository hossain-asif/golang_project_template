package example

import (
	"go_project_structure/common_pkg/logger"
)

var log = logger.Log.Scope("", "middleware", "example_middleware")

// middleware to validate upload csv file
