package file_system

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"go_project_structure/common_pkg/logger"
	"io"
	"os"
	"sync"
	"time"
)

// global variable declaration
var fileSystemLog = logger.Log.Scope("storage", "file_system", "reader")

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
	mu           sync.RWMutex
	index        map[string]Index // map of record IDs to byte offsets
	file         *os.File
	lastModTime  time.Time
	lastChecksum []byte // stores last known file checksum
}

// NewFileStore opens the file and builds the index
func NewFileStore(path string) (*FileStore, error) {
	log := fileSystemLog.Method("NewFileStore")
	f, err := os.Open(path)
	if err != nil {
		log.Errorf("Failed to open file. %v", err)
		return nil, err
	}

	// capture initial mod time
	stat, err := f.Stat()
	if err != nil {
		log.Errorf("Failed to stat file. %v", err)
		return nil, err
	}

	fs := &FileStore{
		index:       make(map[string]Index),
		file:        f,
		lastModTime: stat.ModTime(),
	}

	// compute and store initial checksum
	checksum, err := fs.computeChecksum()
	if err != nil {
		log.Errorf("Failed to compute checksum. %v", err)
		return nil, err
	}
	fs.lastChecksum = checksum

	if err := fs.BuildIndex(); err != nil {
		log.Errorf("Failed to build index. %v", err)
		return nil, err
	}

	return fs, nil
}

func (fs *FileStore) computeChecksum() ([]byte, error) {
	log := fileSystemLog.Method("computeChecksum")
	fs.file.Seek(0, 0) // reset to beginning

	h := md5.New()
	if _, err := io.Copy(h, fs.file); err != nil {
		log.Errorf("Failed to compute checksum. %v", err)
		return nil, err
	}

	return h.Sum(nil), nil
}

func (fs *FileStore) RebuildIfChanged() error {
	log := fileSystemLog.Method("RebuildIfChanged")
	// get current file mod time
	stat, err := fs.file.Stat()
	if err != nil {
		log.Errorf("Failed to stat file. %v", err)
		return err
	}
	currentModTime := stat.ModTime()

	fs.mu.RLock()
	lastMod := fs.lastModTime
	fs.mu.RUnlock()

	// file unchanged — skip rebuild
	if currentModTime.Equal(lastMod) {
		log.Infof("File unchanged.")
		return nil
	}

	// file changed — rebuild
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.index = make(map[string]Index)
	if err := fs.buildIndex(); err != nil {
		log.Errorf("Failed to rebuild index. %v", err)
		return err
	}

	fs.lastModTime = currentModTime // update tracked time
	return nil
}

func (fs *FileStore) RebuildIfCheckSumChanged() error {
	log := fileSystemLog.Method("RebuildIfCheckSumChanged")
	// compute current file checksum
	currentChecksum, err := fs.computeChecksum()
	if err != nil {
		log.Errorf("Failed to compute checksum. %v", err)
		return err
	}

	fs.mu.RLock()
	lastChecksum := fs.lastChecksum
	fs.mu.RUnlock()

	// checksum same — file unchanged, skip rebuild
	if bytes.Equal(currentChecksum, lastChecksum) {
		log.Infof("File checksum unchanged.")
		return nil
	}

	// checksum different — file changed, rebuild
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.index = make(map[string]Index)
	if err := fs.buildIndex(); err != nil {
		log.Errorf("Failed to rebuild index. %v", err)
		return err
	}

	fs.lastChecksum = currentChecksum // update stored checksum
	return nil
}

// buildIndex scans the file ONCE and stores byte offsets
func (fs *FileStore) buildIndex() error {
	log := fileSystemLog.Method("buildIndex")
	fs.file.Seek(0, 0) // reset to beginning

	decoder := json.NewDecoder(fs.file)

	// Read opening '[' of the JSON array
	if _, err := decoder.Token(); err != nil {
		log.Errorf("Failed to read opening '[' of the JSON array. %v", err)
		return err
	}

	for decoder.More() {
		startOffset := decoder.InputOffset()

		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			log.Errorf("Failed to decode record. %v", err)
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
	log.Infof("Index built.")
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
	log := fileSystemLog.Method("GetRaw")
	fs.mu.RLock()
	idx, ok := fs.index[id]
	fs.mu.RUnlock()

	if !ok {
		log.Warnf("Record not found: %s", id)
		return nil, errors.New("record not found: " + id)
	}

	buf := make([]byte, idx.Length)
	_, err := fs.file.ReadAt(buf, idx.Offset)
	if err != nil {
		log.Errorf("Failed to read record. %v", err)
		return nil, err
	}

	buf = bytes.TrimLeft(buf, ",\n\r\t ")
	buf = bytes.TrimRight(buf, ",\n\r\t ")

	return buf, nil
}

// Get parses and returns the record as a map
func (fs *FileStore) GetRecord(id string, dest interface{}) error {
	log := fileSystemLog.Method("GetRecord")
	raw, err := fs.GetRaw(id)
	if err != nil {
		log.Errorf("Failed to get record. %v", err)
		return err
	}
	return json.Unmarshal(raw, dest)
}

// Close cleans up the file handle
func (fs *FileStore) Close() {
	fs.file.Close()
}
