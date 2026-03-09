package app

import (
	"fmt"
	"go_project_structure/common_pkg/storage"
	config "go_project_structure/config/env"
)

// setupFileStore initialises the KV file store from environment config.
func SetupFileStore() (*storage.FileStore, error) {
	dir := config.GetString("KV_FILE_DIRECTORY", "")
	fs, err := storage.NewFileStore(dir, storage.FormatKV)
	if err != nil {
		appLog.Errorf("Failed to build file store: %v", err)
		return nil, fmt.Errorf("file store setup: %w", err)
	}
	return fs, nil
}
