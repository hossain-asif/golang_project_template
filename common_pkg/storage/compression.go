package storage

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Codec compresses and decompresses raw byte payloads.
// It has no knowledge of JSON, records, files, or indexes.
type Codec interface {
	Compress(src []byte) ([]byte, error)
	Decompress(src []byte) ([]byte, error)
}

// NoopCodec passes bytes through unchanged. Used by FormatJSON.
type NoopCodec struct{}

func (NoopCodec) Compress(src []byte) ([]byte, error)   { return src, nil }
func (NoopCodec) Decompress(src []byte) ([]byte, error) { return src, nil }

// GzipCodec compresses with gzip. Used by FormatGzip.
type GzipCodec struct{}

func (GzipCodec) Compress(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(src); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (GzipCodec) Decompress(src []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}