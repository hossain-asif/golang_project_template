package storage

import (
	"crypto/md5"
	"encoding/json"
	"go_project_structure/common_pkg/logger"
	"io"
	"sync"
	"time"
)

var indexerLog = logger.Log.Scope("storage", "file_system", "indexer")

type Index struct {
	Offset int64
	Length int
}

// Indexer maintains the id→Index map and owns change-detection state.
// It does not know about compression or JSON structure — it delegates
// scanning to the Storage it is given.
type Indexer struct {
	mu           sync.RWMutex
	index        map[string]Index
	lastModTime  time.Time
	lastChecksum []byte
	storage      Storage
}

func NewIndexer(storage Storage) (*Indexer, error) {
	log := indexerLog.Method("NewIndexer")

	stat, err := storage.File().Stat()
	if err != nil {
		log.Errorf("Failed to stat file. %v", err)
		return nil, err
	}

	ix := &Indexer{
		index:   make(map[string]Index),
		storage: storage,
	}

	ix.lastModTime = stat.ModTime()

	checksum, err := ix.computeChecksum()
	if err != nil {
		log.Errorf("Failed to compute initial checksum. %v", err)
		return nil, err
	}
	ix.lastChecksum = checksum

	if err := ix.Rebuild(); err != nil {
		log.Errorf("Failed to build initial index. %v", err)
		return nil, err
	}

	return ix, nil
}

// Lookup returns the Index for id, and whether it was found.
func (ix *Indexer) Lookup(id string) (Index, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	idx, ok := ix.index[id]
	return idx, ok
}

// Set inserts or updates a single entry without a full rebuild.
// Called by FileStore after a successful Append.
func (ix *Indexer) Set(id string, idx Index) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.index[id] = idx
}

// Rebuild scans the backing storage and rebuilds the full id→Index map.
func (ix *Indexer) Rebuild() error {
	log := indexerLog.Method("Rebuild")

	newIndex := make(map[string]Index)

	err := ix.storage.Scan(func(idx Index, payload []byte) error {
		var record map[string]interface{}
		if err := json.Unmarshal(payload, &record); err != nil {
			log.Warnf("Skipping unparseable record at offset %d: %v", idx.Offset, err)
			return nil // skip bad records, don't abort the scan
		}
		if id, ok := record["id"].(string); ok {
			newIndex[id] = idx
		}
		return nil
	})
	if err != nil {
		log.Errorf("Scan failed. %v", err)
		return err
	}

	ix.mu.Lock()
	ix.index = newIndex
	ix.mu.Unlock()

	log.Infof("Index rebuilt with %d records.", len(newIndex))
	return nil
}

// RebuildIfChanged checks the file mod-time and rebuilds only if it changed.
func (ix *Indexer) RebuildIfChanged() error {
	log := indexerLog.Method("RebuildIfChanged")

	stat, err := ix.storage.File().Stat()
	if err != nil {
		log.Errorf("Failed to stat file. %v", err)
		return err
	}

	ix.mu.RLock()
	unchanged := stat.ModTime().Equal(ix.lastModTime)
	ix.mu.RUnlock()

	if unchanged {
		log.Infof("File unchanged (mod-time).")
		return nil
	}

	if err := ix.Rebuild(); err != nil {
		return err
	}

	ix.mu.Lock()
	ix.lastModTime = stat.ModTime()
	ix.mu.Unlock()

	return nil
}

// RebuildIfChecksumChanged computes the MD5 of the file and rebuilds only if it changed.
func (ix *Indexer) RebuildIfChecksumChanged() error {
	log := indexerLog.Method("RebuildIfChecksumChanged")

	current, err := ix.computeChecksum()
	if err != nil {
		log.Errorf("Failed to compute checksum. %v", err)
		return err
	}

	ix.mu.RLock()
	same := checksumEqual(current, ix.lastChecksum)
	ix.mu.RUnlock()

	if same {
		log.Infof("File unchanged (checksum).")
		return nil
	}

	if err := ix.Rebuild(); err != nil {
		return err
	}

	ix.mu.Lock()
	ix.lastChecksum = current
	ix.mu.Unlock()

	return nil
}

// UpdateChecksum recomputes and stores the checksum.
// Called by FileStore after a write so the next RebuildIfChecksumChanged
// does not trigger a spurious rebuild.
func (ix *Indexer) UpdateChecksum() error {
	checksum, err := ix.computeChecksum()
	if err != nil {
		return err
	}
	ix.mu.Lock()
	ix.lastChecksum = checksum
	ix.mu.Unlock()
	return nil
}

func (ix *Indexer) computeChecksum() ([]byte, error) {
	f := ix.storage.File()
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func checksumEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
