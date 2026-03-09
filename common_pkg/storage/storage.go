package storage

import (
	"os"
)

type Storage interface {
	// ReadAt reads the raw (possibly compressed) payload at the given Index.
	ReadAt(idx Index) ([]byte, error)

	// Append writes a new payload to the store and returns the Index
	// that was assigned to it (so the caller can update the index).
	Append(payload []byte) (Index, error)

	// Scan calls visit for every payload in the file, in order.
	// Used by the Indexer to rebuild the id→Index map.
	Scan(visit func(idx Index, payload []byte) error) error

	// File returns the underlying *os.File (used for checksum, stat, reopen).
	File() *os.File

	// ReplaceFile is called after a full file rewrite (FormatJSON only).
	// It swaps the internal file handle for the newly opened one.
	ReplaceFile(f *os.File)
}
