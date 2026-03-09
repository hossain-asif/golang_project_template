package storage

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// KVStorage implements the Storage interface using a plain-text key-value file.
//
// On-disk format (one record per line):
//
//   <key>\t<base64(compressed JSON)>\n
//
// The key is plain text — human-readable and greppable with standard tools.
// The value is the JSON payload compressed by the provided Codec, then
// base64-encoded so the file stays valid UTF-8 text throughout.
//
// Example line on disk:
//   u1	H4sIAAAAAAAA/6pWKkktLlGyUlIqS...==
//
// Index.Offset = byte offset of the first character of the line (the key).
// Index.Length = total byte length of the line including the trailing '\n'.
//
// Lookup:  O(1) — ReadAt seeks directly to the line by offset.
// Append:  O(1) — new lines are appended at EOF; no rewrite needed.
// Update:  rewrites the file (marks old line as deleted, appends new).
// Delete:  rewrites the file omitting the deleted key.
type KVStorage struct {
	file  *os.File
	codec Codec
}

func NewKVStorage(path string, codec Codec) (*KVStorage, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &KVStorage{file: f, codec: codec}, nil
}

func (s *KVStorage) File() *os.File      { return s.file }
func (s *KVStorage) ReplaceFile(f *os.File) { s.file = f }

// ReadAt reads the line at idx, strips the key prefix, base64-decodes,
// then decompresses and returns the raw JSON value.
func (s *KVStorage) ReadAt(idx Index) ([]byte, error) {
	line := make([]byte, idx.Length)
	if _, err := s.file.ReadAt(line, idx.Offset); err != nil {
		return nil, err
	}

	// Strip trailing newline.
	line = bytes.TrimRight(line, "\n")

	// Split on the tab separator between key and value.
	tab := bytes.IndexByte(line, '\t')
	if tab == -1 {
		return nil, errors.New("kv: malformed line — no tab separator found")
	}

	encoded := line[tab+1:]

	compressed, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, fmt.Errorf("kv: base64 decode failed: %w", err)
	}

	return s.codec.Decompress(compressed)
}

// Append compresses payload, base64-encodes it, and appends a new
// "<key>\t<encoded>\n" line at EOF.
//
// The key must be provided as the "id" field inside the JSON payload.
// Call extractKey() before Append if you need the key separately.
func (s *KVStorage) Append(payload []byte) (Index, error) {
	key, err := extractKey(payload)
	if err != nil {
		return Index{}, err
	}
	return s.appendKV(key, payload)
}

// AppendKV lets callers specify the key explicitly, independent of the
// JSON payload content. Useful when the key is not the "id" field.
func (s *KVStorage) AppendKV(key string, payload []byte) (Index, error) {
	return s.appendKV(key, payload)
}

func (s *KVStorage) appendKV(key string, payload []byte) (Index, error) {
	if strings.ContainsAny(key, "\t\n") {
		return Index{}, errors.New("kv: key must not contain tab or newline characters")
	}

	compressed, err := s.codec.Compress(payload)
	if err != nil {
		return Index{}, err
	}

	encoded := base64.StdEncoding.EncodeToString(compressed)
	line := key + "\t" + encoded + "\n"

	writeOffset, err := s.file.Seek(0, io.SeekEnd)
	if err != nil {
		return Index{}, err
	}

	if _, err := s.file.WriteString(line); err != nil {
		return Index{}, err
	}

	if err := s.file.Sync(); err != nil {
		return Index{}, err
	}

	return Index{Offset: writeOffset, Length: len(line)}, nil
}

// Scan reads the file line by line and calls visit for each valid key-value line.
// The payload passed to visit is already decompressed JSON.
// Blank lines and lines not containing a tab are silently skipped.
func (s *KVStorage) Scan(visit func(idx Index, payload []byte) error) error {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	scanner := bufio.NewScanner(s.file)

	// Increase scanner buffer for large compressed+encoded values.
	buf := make([]byte, 0, 1*1024*1024) // 1 MB initial
	scanner.Buffer(buf, 64*1024*1024)   // up to 64 MB per line

	var offset int64

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		lineLen := int64(len(lineBytes)) + 1 // +1 for the '\n' bufio strips

		// Skip blank lines.
		if len(lineBytes) == 0 {
			offset += lineLen
			continue
		}

		tab := bytes.IndexByte(lineBytes, '\t')
		if tab == -1 {
			// Not a valid KV line — skip.
			offset += lineLen
			continue
		}

		encoded := lineBytes[tab+1:]

		compressed, err := base64.StdEncoding.DecodeString(string(encoded))
		if err != nil {
			// Corrupted line — skip, don't abort the whole scan.
			offset += lineLen
			continue
		}

		raw, err := s.codec.Decompress(compressed)
		if err != nil {
			offset += lineLen
			continue
		}

		idx := Index{Offset: offset, Length: int(lineLen)}
		if err := visit(idx, raw); err != nil {
			return err
		}

		offset += lineLen
	}

	return scanner.Err()
}

// UpdateKV overwrites the value for an existing key.
// It rewrites the entire file, replacing the old line with the new value.
// If the key does not exist, it appends a new line instead.
func (s *KVStorage) UpdateKV(key string, payload []byte) (Index, error) {
	compressed, err := s.codec.Compress(payload)
	if err != nil {
		return Index{}, err
	}
	encoded := base64.StdEncoding.EncodeToString(compressed)
	newLine := key + "\t" + encoded + "\n"

	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return Index{}, err
	}

	content, err := io.ReadAll(s.file)
	if err != nil {
		return Index{}, err
	}

	var out bytes.Buffer
	var recordOffset int64 = -1
	var currentOffset int64

	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 1*1024*1024), 64*1024*1024)

	found := false
	for scanner.Scan() {
		line := scanner.Text()
		lineLen := int64(len(line)) + 1

		tab := strings.IndexByte(line, '\t')
		if tab != -1 && line[:tab] == key {
			recordOffset = int64(out.Len())
			out.WriteString(newLine)
			found = true
		} else {
			out.WriteString(line + "\n")
		}
		currentOffset += lineLen
	}

	if !found {
		// Key not found — append.
		recordOffset = int64(out.Len())
		out.WriteString(newLine)
	}

	filePath := s.file.Name()
	s.file.Close()

	if err := os.WriteFile(filePath, out.Bytes(), 0644); err != nil {
		return Index{}, err
	}

	s.file, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return Index{}, err
	}

	return Index{Offset: recordOffset, Length: len(newLine)}, nil
}

// DeleteKV removes the line for key and rewrites the file.
func (s *KVStorage) DeleteKV(key string) error {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	content, err := io.ReadAll(s.file)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 1*1024*1024), 64*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		tab := strings.IndexByte(line, '\t')
		if tab != -1 && line[:tab] == key {
			continue // skip this line
		}
		out.WriteString(line + "\n")
	}

	filePath := s.file.Name()
	s.file.Close()

	if err := os.WriteFile(filePath, out.Bytes(), 0644); err != nil {
		return err
	}

	s.file, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	return err
}

// extractKey pulls the "id" string field from a JSON payload.
func extractKey(payload []byte) (string, error) {
	// Fast path: scan for "id":"..." without full unmarshal.
	// Falls back to empty string if not found.
	prefix := []byte(`"id":"`)
	idx := bytes.Index(payload, prefix)
	if idx == -1 {
		return "", errors.New("kv: payload has no \"id\" field")
	}
	start := idx + len(prefix)
	end := bytes.IndexByte(payload[start:], '"')
	if end == -1 {
		return "", errors.New("kv: malformed \"id\" field in payload")
	}
	return string(payload[start : start+end]), nil
}