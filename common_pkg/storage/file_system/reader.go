package file_system

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

var defaultStore *FileStore

// SetDefault sets the global file store instance
func SetDefault(fs *FileStore) {
	defaultStore = fs
}

type Index struct {
	Offset int64
	Length int
}

type FileStore struct {
	mu    sync.RWMutex
	index map[string]Index //
	file  *os.File
}

// NewFileStore opens the file and builds the index
func NewFileStore(path string) (*FileStore, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fs := &FileStore{
		index: make(map[string]Index),
		file:  f,
	}

	if err := fs.BuildIndex(); err != nil {
		return nil, err
	}

	return fs, nil
}

// buildIndex scans the file ONCE and stores byte offsets
func (fs *FileStore) buildIndex() error {
	fs.file.Seek(0, 0) // reset to beginning

	decoder := json.NewDecoder(fs.file)

	// Read opening '[' of the JSON array
	if _, err := decoder.Token(); err != nil {
		return err
	}

	for decoder.More() {
		startOffset := decoder.InputOffset()

		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			return err
		}

		// Your unique key field — change "id" to your actual key
		id, ok := record["id"].(string)
		if !ok {
			continue
		}

		endOffset := decoder.InputOffset()
		fs.index[id] = Index{
			Offset: startOffset,
			Length: int(endOffset - startOffset),
		}
	}
	return nil
}


// public — acquires lock itself
func (fs *FileStore) BuildIndex() error {
    fs.mu.Lock()
    defer fs.mu.Unlock()
    return fs.buildIndex()
}


// GetRaw returns raw JSON bytes for a given id
func (fs *FileStore) GetRaw(id string) ([]byte, error) {
	fs.mu.RLock()
	idx, ok := fs.index[id]
	fs.mu.RUnlock()

	if !ok {
		return nil, errors.New("record not found: " + id)
	}

	buf := make([]byte, idx.Length)
	_, err := fs.file.ReadAt(buf, idx.Offset)
	if err != nil {
		return nil, err
	}

	buf = bytes.TrimLeft(buf, ",\n\r\t ")
	buf = bytes.TrimRight(buf, ",\n\r\t ")

	return buf, err
}

// Get parses and returns the record as a map
func (fs *FileStore) GetRecord(id string, dest interface{}) error {
	raw, err := fs.GetRaw(id)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

// Close cleans up the file handle
func (fs *FileStore) Close() {
	fs.file.Close()
}
