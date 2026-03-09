package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"go_project_structure/common_pkg/logger"
)

var storeLog = logger.Log.Scope("storage", "file_system", "store")

var defaultStore *FileStore

func SetDefault(fs *FileStore) {
	defaultStore = fs
}

// StoreFormat selects the on-disk layout used by a FileStore.
type StoreFormat int

const (
	FormatJSON StoreFormat = iota // plain JSON array, human-readable
	FormatGzip                    // binary [4-byte length][gzip payload]
	FormatKV                      // plain-text <key>\t<base64(gzip(JSON))>\n
)

// FileStore is the public entry point.
// It owns no file handles, no compression logic, and no index state —
// it only orchestrates Storage (I/O) and Indexer (change detection + lookup).
type FileStore struct {
	storage Storage
	indexer *Indexer
}

// NewFileStore constructs a FileStore for the given path and format.
func NewFileStore(path string, format StoreFormat) (*FileStore, error) {
	log := storeLog.Method("NewFileStore")

	var (
		storage Storage
		err     error
	)

	switch format {
	case FormatJSON:
		storage, err = NewJSONStorage(path)
	case FormatGzip:
		storage, err = NewBinaryStorage(path, GzipCodec{})
	case FormatKV:
		storage, err = NewKVStorage(path, GzipCodec{})
	default:
		return nil, errors.New("unknown store format")
	}
	if err != nil {
		log.Errorf("Failed to initialise storage. %v", err)
		return nil, err
	}

	indexer, err := NewIndexer(storage)
	if err != nil {
		log.Errorf("Failed to initialise indexer. %v", err)
		return nil, err
	}

	return &FileStore{storage: storage, indexer: indexer}, nil
}

// AddRecord serialises record to JSON and appends it to the store.
func (fs *FileStore) AddRecord(record interface{}) error {
	log := storeLog.Method("AddRecord")

	payload, err := json.Marshal(record)
	if err != nil {
		log.Errorf("Failed to marshal record. %v", err)
		return err
	}

	idx, err := fs.storage.Append(payload)
	if err != nil {
		log.Errorf("Failed to append record. %v", err)
		return err
	}

	if id := PeekKey(payload); id != "" {
		fs.indexer.Set(id, idx)
	}

	if err := fs.indexer.UpdateChecksum(); err != nil {
		log.Warnf("Failed to update checksum after write. %v", err)
	}

	return nil
}

// UpdateRecord replaces the record for id with the new value.
//
//	FormatJSON — splices the new JSON in place by byte offset, then rebuilds index.
//	FormatKV   — rewrites the matching line, then rebuilds index.
//	FormatGzip — not supported (binary format is append-only).
func (fs *FileStore) UpdateRecord(id string, record interface{}) error {
	log := storeLog.Method("UpdateRecord")

	payload, err := json.Marshal(record)
	if err != nil {
		log.Errorf("Failed to marshal record. %v", err)
		return err
	}

	switch s := fs.storage.(type) {

	case *JSONStorage:
		idx, ok := fs.indexer.Lookup(id)
		if !ok {
			return errors.New("record not found: " + id)
		}
		newIdx, err := s.UpdateJSON(idx, payload)
		if err != nil {
			log.Errorf("Failed to update JSON record. %v", err)
			return err
		}
		// Offsets of records after this one may have shifted — full rebuild.
		if err := fs.indexer.Rebuild(); err != nil {
			log.Errorf("Failed to rebuild index after update. %v", err)
			return err
		}
		// Re-set the updated record with its fresh offset.
		fs.indexer.Set(id, newIdx)

	case *KVStorage:
		newIdx, err := s.UpdateKV(id, payload)
		if err != nil {
			log.Errorf("Failed to update KV record. %v", err)
			return err
		}
		if err := fs.indexer.Rebuild(); err != nil {
			log.Errorf("Failed to rebuild index after KV update. %v", err)
			return err
		}
		fs.indexer.Set(id, newIdx)

	default:
		return errors.New("UpdateRecord is not supported for FormatGzip")
	}

	if err := fs.indexer.UpdateChecksum(); err != nil {
		log.Warnf("Failed to update checksum after update. %v", err)
	}

	return nil
}

// DeleteRecord removes the record with the given id from the store.
//
//	FormatJSON — cuts the record out of the array by byte offset, then rebuilds index.
//	FormatKV   — removes the matching line, then rebuilds index.
//	FormatGzip — not supported (binary format is append-only).
func (fs *FileStore) DeleteRecord(id string) error {
	log := storeLog.Method("DeleteRecord")

	switch s := fs.storage.(type) {

	case *JSONStorage:
		idx, ok := fs.indexer.Lookup(id)
		if !ok {
			return errors.New("record not found: " + id)
		}
		if err := s.DeleteJSON(idx); err != nil {
			log.Errorf("Failed to delete JSON record. %v", err)
			return err
		}
		// All offsets after the deleted record have shifted — full rebuild.
		if err := fs.indexer.Rebuild(); err != nil {
			log.Errorf("Failed to rebuild index after delete. %v", err)
			return err
		}

	case *KVStorage:
		if err := s.DeleteKV(id); err != nil {
			log.Errorf("Failed to delete KV record. %v", err)
			return err
		}
		if err := fs.indexer.Rebuild(); err != nil {
			log.Errorf("Failed to rebuild index after KV delete. %v", err)
			return err
		}

	default:
		return errors.New("DeleteRecord is not supported for FormatGzip")
	}

	if err := fs.indexer.UpdateChecksum(); err != nil {
		log.Warnf("Failed to update checksum after delete. %v", err)
	}

	return nil
}

// GetRaw returns the decompressed JSON bytes for the given id.
func (fs *FileStore) GetRaw(id string) ([]byte, error) {
	log := storeLog.Method("GetRaw")

	idx, ok := fs.indexer.Lookup(id)
	if !ok {
		log.Warnf("Record not found: %s", id)
		return nil, errors.New("record not found: " + id)
	}

	raw, err := fs.storage.ReadAt(idx)
	if err != nil {
		log.Errorf("Failed to read record %s. %v", id, err)
		return nil, err
	}

	raw = bytes.TrimLeft(raw, ",\n\r\t ")
	raw = bytes.TrimRight(raw, ",\n\r\t ")

	return raw, nil
}

// GetRecord deserialises the record for id into dest.
func (fs *FileStore) GetRecord(id string, dest interface{}) error {
	log := storeLog.Method("GetRecord")

	raw, err := fs.GetRaw(id)
	if err != nil {
		log.Errorf("Failed to get record. %v", err)
		return err
	}

	return json.Unmarshal(raw, dest)
}

// RebuildIfChanged delegates change-detection (by mod-time) to the Indexer.
func (fs *FileStore) RebuildIfChanged() error {
	return fs.indexer.RebuildIfChanged()
}

// RebuildIfChecksumChanged delegates change-detection (by MD5) to the Indexer.
func (fs *FileStore) RebuildIfChecksumChanged() error {
	return fs.indexer.RebuildIfChecksumChanged()
}

// BuildIndex forces a full index rebuild regardless of change state.
func (fs *FileStore) BuildIndex() error {
	return fs.indexer.Rebuild()
}

// Close releases the underlying file handle.
func (fs *FileStore) Close() {
	fs.storage.File().Close()
}
