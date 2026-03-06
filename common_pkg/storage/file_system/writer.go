package file_system

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
)

// AddRecord appends a new record to the JSON array in the file
func (fs *FileStore) AddRecord(record interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Marshal the new record to JSON
	newData, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Get current file size
	stat, err := fs.file.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()

	// Read entire file to find the closing ']'
	raw := make([]byte, fileSize)
	_, err = fs.file.ReadAt(raw, 0)
	if err != nil {
		return err
	}

	// Find last ']' position
	closingBracket := bytes.LastIndex(raw, []byte("]"))
	if closingBracket == -1 {
		return errors.New("invalid JSON file: no closing bracket found")
	}

	// Find last '}' to determine if array is empty or has records
	lastRecord := bytes.LastIndex(raw[:closingBracket], []byte("}"))

	var finalContent []byte
	if lastRecord == -1 {
		// Empty array [] → insert first record
        finalContent = append(finalContent, raw[:closingBracket]...)
        finalContent = append(finalContent, '\n')
        finalContent = append(finalContent, newData...)
        finalContent = append(finalContent, '\n')
	} else {
       // comma goes RIGHT after last '}', then newline, then new record
        finalContent = append(finalContent, raw[:lastRecord+1]...) // up to last '}'
        finalContent = append(finalContent, ',')                  // comma here
        finalContent = append(finalContent, '\n')
        finalContent = append(finalContent, newData...)
        finalContent = append(finalContent, '\n')
	}
	finalContent = append(finalContent, ']')

	// Reopen file in write mode and overwrite
	filePath := fs.file.Name()
	fs.file.Close()

	err = os.WriteFile(filePath, finalContent, 0644)
	if err != nil {
		return err
	}

	// Reopen for reading
	fs.file, err = os.Open(filePath)
	if err != nil {
		return err
	}

    // update checksum after write so scheduler doesn't rebuild unnecessarily
    checksum, err := fs.computeChecksum()
    if err != nil {
        return err
    }
    fs.lastChecksum = checksum

	// Rebuild index to include new record
	// fs.index = make(map[string]Index)
	// return fs.buildIndex()
	return nil
}
