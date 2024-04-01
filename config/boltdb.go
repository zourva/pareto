package config

import (
	"errors"
	"path/filepath"
)

// BoltdbProvider implements a boltdb provider.
// To use the provider correctly, a concrete parser
// is needed to parse raw bytes returned by ReadBytes.
type BoltdbProvider struct {
	path   string // path to db file
	cfgTab string // table name of config
	cfgKey string // key of config value

	storage *storage
}

// NewBoltdbProvider returns a boltdb provider.
func NewBoltdbProvider(path string) *BoltdbProvider {
	file := filepath.Clean(path)

	return &BoltdbProvider{
		path:    file,
		cfgTab:  "global",
		cfgKey:  "config",
		storage: newStorage(file),
	}
}

// ReadBytes is not supported by boltdb provider.
func (f *BoltdbProvider) ReadBytes() ([]byte, error) {
	err := f.storage.open()
	if err != nil {
		return nil, err
	}

	defer f.storage.close()

	bytes, err := f.storage.queryOne(f.cfgTab, f.cfgKey)
	if err != nil {
		return nil, err
	}

	// deep copy, otherwise bytes is invalid after f.storage.close()
	var value = make([]byte, len(bytes))
	copy(value, bytes)

	return value, nil
}

// Read is not supported by boltdb provider.
func (f *BoltdbProvider) Read() (map[string]any, error) {
	return nil, errors.New("boltdb provider does not support this method")
}
