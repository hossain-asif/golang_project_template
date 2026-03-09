package storage

import "encoding/json"

func lastIndex(b []byte, ch byte) int {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == ch {
			return i
		}
	}
	return -1
}

// peekID extracts the "id" string field from a marshalled JSON payload.
func PeekKey(payload []byte) string {
	var peek map[string]interface{}
	if err := json.Unmarshal(payload, &peek); err != nil {
		return ""
	}
	id, _ := peek["id"].(string)
	return id
}
