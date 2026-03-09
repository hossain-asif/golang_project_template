package storage

import (
	"errors"
	"io"
	"os"
)

type JSONStorage struct {
	file *os.File
}

func NewJSONStorage(path string) (*JSONStorage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONStorage{file: file}, nil
}

func (s *JSONStorage) File() *os.File         { return s.file }
func (s *JSONStorage) ReplaceFile(f *os.File) { s.file = f }

// ReadAt reads the raw JSON bytes for the record described by idx.
func (s *JSONStorage) ReadAt(idx Index) ([]byte, error) {
	buf := make([]byte, idx.Length)
	if _, err := s.file.ReadAt(buf, idx.Offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// Append rewrites the JSON array to include the new payload at the end.
// Returns the Index pointing at the new record's position in the rewritten file.
func (s *JSONStorage) Append(payload []byte) (Index, error) {
	content, err := s.readAll()
	if err != nil {
		return Index{}, err
	}

	closingBracket := lastIndex(content, ']')
	if closingBracket == -1 {
		return Index{}, errors.New("json: invalid file — no closing ']' found")
	}

	lastBrace := lastIndex(content[:closingBracket], '}')

	var final []byte
	var recordOffset int64

	if lastBrace == -1 {
		// Empty array: [\n<payload>\n]
		final = append(final, content[:closingBracket]...)
		final = append(final, '\n')
		recordOffset = int64(len(final))
		final = append(final, payload...)
		final = append(final, '\n')
	} else {
		// Non-empty: ...},\n<payload>\n]
		final = append(final, content[:lastBrace+1]...)
		final = append(final, ',', '\n')
		recordOffset = int64(len(final))
		final = append(final, payload...)
		final = append(final, '\n')
	}
	final = append(final, ']')

	if err := s.rewrite(final); err != nil {
		return Index{}, err
	}

	return Index{Offset: recordOffset, Length: len(payload)}, nil
}

// UpdateJSON replaces the record at idx with newPayload.
// The record slice (idx.Offset to idx.Offset+idx.Length) is cut out and
// newPayload is spliced in — no other records are touched.
// Returns the updated Index for the replaced record.
func (s *JSONStorage) UpdateJSON(idx Index, newPayload []byte) (Index, error) {
	content, err := s.readAll()
	if err != nil {
		return Index{}, err
	}

	start := idx.Offset
	end := idx.Offset + int64(idx.Length)

	if start < 0 || end > int64(len(content)) {
		return Index{}, errors.New("json: update index is out of bounds")
	}

	// Verify we are actually pointing at a JSON object.
	if content[start] != '{' {
		return Index{}, errors.New("json: offset does not point to a '{' — stale index?")
	}

	// Build: everything before + newPayload + everything after.
	var final []byte
	final = append(final, content[:start]...)
	recordOffset := int64(len(final))
	final = append(final, newPayload...)
	final = append(final, content[end:]...)

	if err := s.rewrite(final); err != nil {
		return Index{}, err
	}

	return Index{Offset: recordOffset, Length: len(newPayload)}, nil
}

// DeleteJSON removes the record at idx from the JSON array.
// It also removes the surrounding comma and whitespace so the array
// stays valid JSON after the deletion.
func (s *JSONStorage) DeleteJSON(idx Index) error {
	content, err := s.readAll()
	if err != nil {
		return err
	}

	start := idx.Offset
	end := idx.Offset + int64(idx.Length)

	if start < 0 || end > int64(len(content)) {
		return errors.New("json: delete index is out of bounds")
	}

	if content[start] != '{' {
		return errors.New("json: offset does not point to a '{' — stale index?")
	}

	// --- trim the comma and whitespace that bind this record to its neighbours ---
	//
	// Four possible layouts around the target record T:
	//
	//   (a) Only record:      [ \n T \n ]          → leave just [ \n ]
	//   (b) First of many:    [ \n T ,\n N ... ]    → remove T and the trailing comma
	//   (c) Last of many:     [ ... P ,\n T \n ]    → remove the leading comma after P
	//   (d) Middle:           [ ... P ,\n T ,\n N ] → remove T and one surrounding comma

	// Expand left to include any leading whitespace/comma before the record.
	trimStart := start
	for trimStart > 0 && isWhitespaceOrComma(content[trimStart-1]) {
		trimStart--
	}

	// Expand right to include any trailing comma and whitespace after the record.
	trimEnd := end
	for trimEnd < int64(len(content)) && isWhitespaceOrComma(content[trimEnd]) {
		trimEnd++
	}

	// After trimming, make sure at least one '\n' separates the array bracket
	// from the next record (if any remain), so the file stays readable.
	var final []byte
	before := content[:trimStart]
	after := content[trimEnd:]

	final = append(final, before...)

	// If there are remaining records, restore a newline separator.
	if len(after) > 0 && after[0] != ']' {
		final = append(final, '\n')
	} else if len(after) > 0 && after[0] == ']' {
		// No records left — keep a clean newline before the closing bracket.
		final = append(final, '\n')
	}

	final = append(final, after...)

	return s.rewrite(final)
}

// Scan calls visit for every JSON object in the array.
// Byte offsets are exact so the returned Index values are valid for ReadAt,
// UpdateJSON, and DeleteJSON.
func (s *JSONStorage) Scan(visit func(idx Index, payload []byte) error) error {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	content, err := io.ReadAll(s.file)
	if err != nil {
		return err
	}

	i := 0
	for i < len(content) {
		if content[i] != '{' {
			i++
			continue
		}

		start := i
		depth, inString := 0, false

		for i < len(content) {
			ch := content[i]
			if inString {
				if ch == '\\' {
					i += 2
					continue
				}
				if ch == '"' {
					inString = false
				}
			} else {
				switch ch {
				case '"':
					inString = true
				case '{':
					depth++
				case '}':
					depth--
					if depth == 0 {
						end := i + 1
						err := visit(
							Index{Offset: int64(start), Length: end - start},
							content[start:end],
						)
						if err != nil {
							return err
						}
						i = end
						goto nextRecord
					}
				}
			}
			i++
		}
	nextRecord:
	}

	return nil
}

// ---- internal helpers ----

// readAll reads the entire file content into memory.
func (s *JSONStorage) readAll() ([]byte, error) {
	stat, err := s.file.Stat()
	if err != nil {
		return nil, err
	}

	// Brand new / empty file — seed it as an empty JSON array.
	if stat.Size() == 0 {
		return []byte("[]"), nil
	}

	buf := make([]byte, stat.Size())
	if _, err := s.file.ReadAt(buf, 0); err != nil && err != io.EOF {
		return nil, err
	}
	return buf, nil
}

// rewrite atomically replaces the file content and reopens the handle.
func (s *JSONStorage) rewrite(content []byte) error {
	filePath := s.file.Name()
	s.file.Close()

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return err
	}

	var err error
	s.file, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	return err
}

func isWhitespaceOrComma(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ','
}
