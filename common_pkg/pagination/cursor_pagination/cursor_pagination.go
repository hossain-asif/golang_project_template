package cursor_pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Cursor holds the encoded position in the dataset
type Cursor struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Direction Direction `json:"direction"`
}

// Encode encodes the cursor to a base64 string safe for URLs
func (c *Cursor) EncodeCursor() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("cursor encode: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeCursor decodes a base64 cursor string
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	// Sanitize (strip whitespace that sometimes appears in URLs)
	encoded = strings.TrimSpace(encoded)

	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid cursor format")
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, errors.New("invalid cursor data")
	}
	return &c, nil
}
