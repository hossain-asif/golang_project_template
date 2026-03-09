package storage

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// BinaryStorage implements RandomAccessStorage using a flat binary format
// with per-record gzip (or any Codec) compression.
//
// On-disk layout per record:
//   [4-byte big-endian uint32: compressed length][N bytes: compressed JSON]
//
// Index.Offset = byte offset of the 4-byte length header.
// Index.Length = compressed payload length N (NOT including the 4-byte header).
type BinaryStorage struct {
	file  *os.File
	codec Codec
}

func NewBinaryStorage(path string, codec Codec) (*BinaryStorage, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &BinaryStorage{file: f, codec: codec}, nil
}

func (s *BinaryStorage) File() *os.File { 
	return s.file 
}
func (s *BinaryStorage) ReplaceFile(f *os.File) {
	s.file = f
}

// ReadAt reads and decompresses the payload at idx.
// idx.Offset points at the 4-byte header; the payload starts at Offset+4.
func (s *BinaryStorage) ReadAt(idx Index) ([]byte, error) {
	compressed := make([]byte, idx.Length)
	if _, err := s.file.ReadAt(compressed, idx.Offset+4); err != nil {
		return nil, err
	}
	return s.codec.Decompress(compressed)
}

// Append compresses payload and appends [header][payload] at EOF.
func (s *BinaryStorage) Append(payload []byte) (Index, error) {
	compressed, err := s.codec.Compress(payload)
	if err != nil {
		return Index{}, err
	}
	if len(compressed) > 0xFFFFFFFF {
		return Index{}, errors.New("compressed record exceeds 4 GB limit")
	}

	writeOffset, err := s.file.Seek(0, io.SeekEnd)
	if err != nil {
		return Index{}, err
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(compressed)))

	if _, err := s.file.Write(header[:]); err != nil {
		return Index{}, err
	}
	if _, err := s.file.Write(compressed); err != nil {
		return Index{}, err
	}
	if err := s.file.Sync(); err != nil {
		return Index{}, err
	}

	return Index{Offset: writeOffset, Length: len(compressed)}, nil
}

// Scan visits every record in the binary file in order.
func (s *BinaryStorage) Scan(visit func(idx Index, payload []byte) error) error {
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	var offset int64
	var header [4]byte

	for {
		if _, err := io.ReadFull(s.file, header[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}

		compressedLen := int(binary.BigEndian.Uint32(header[:]))
		compressed := make([]byte, compressedLen)

		if _, err := io.ReadFull(s.file, compressed); err != nil {
			return err
		}

		raw, err := s.codec.Decompress(compressed)
		if err != nil {
			return err
		}

		if err := visit(Index{Offset: offset, Length: compressedLen}, raw); err != nil {
			return err
		}

		offset += 4 + int64(compressedLen)
	}
}