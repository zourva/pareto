package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zourva/pareto/box"
	"io"
	"os"
	"strings"
)

// Store wraps viper and provides extended functionalities.
// Store uses a storage system as the backlog and accepts
// multiple config sources to merge them into the storage.
//
// A config store is structured as a tree and be flattened
// using a key-value pattern, and the config value is accessed
// using a key, depicted by a tree-node-path.
type Store struct {
	*viper.Viper
	rwc io.ReadWriteCloser //backlog
}

// Load loads configuration from the underlying storage.
// The loaded configuration is presented managed as a tree,
// whose root node is identified by rootPaths.
func (s *Store) Load(kind string, rootPaths ...string) error {
	s.SetConfigType(kind)

	rootKey := strings.Join(rootPaths, ".")
	s.Get(rootKey)

	err := s.ReadConfig(s.rwc)
	if err != nil {
		return err
	}

	return nil
}

// NewStore creates a configuration store based on the given config file.
//
// The supported config file content has the same set of extensions provided
// by viper (either one of json/toml/yaml/yml/properties/props/prop/hcl/tfvars
// /dotenv/env/ini or extended type db)
//
// rootPaths, if provided, identifies a subtree, and will be used as the config
// tree root and if the given path does not exist, a new root node is created.
//
// Returns the created store or nil if any error occurred.
func NewStore(file string, rootPaths ...string) (*Store, error) {
	ok, err := box.PathExists(file)
	if err != nil {
		log.Errorf("access file %s failed: %v", file, err)
		return nil, err
	}

	if !ok {
		log.Errorf("file %s does not exist", file)
		return nil, os.ErrNotExist
	}

	v := viper.NewWithOptions(
		viper.KeyDelimiter("."),
	)

	v.EncoderRegistry().RegisterEncoder("db", DBCodec{})
	v.DecoderRegistry().RegisterDecoder("db", DBCodec{})
	v.SetConfigFile(file)

	err = v.ReadInConfig()
	if err != nil {
		log.Errorf("read config file %s failed: %v", file, err)
		return nil, err
	}

	if len(rootPaths) != 0 {
		rootKey := strings.Join(rootPaths, ".")
		newRoot := v.Sub(rootKey)

		// overwrite root iff new root node is valid
		if newRoot != nil {
			v = newRoot
		}
	}

	s := &Store{
		Viper: v,
	}

	return s, nil
}

//func New(kind string, rwc io.ReadWriteCloser, rootPaths ...string) (*Store, error) {
//	if !supported(kind) {
//		return nil, viper.UnsupportedConfigError(kind)
//	}
//
//	if rwc == nil {
//		return nil, errors.New("config buffer must not be nil")
//	}
//
//	v := viper.NewWithOptions(
//		viper.KeyDelimiter("."),
//	)
//
//	v.EncoderRegistry().RegisterEncoder("db", DBCodec{})
//	v.DecoderRegistry().RegisterDecoder("db", DBCodec{})
//	v.SetConfigType(kind)
//
//	err := v.ReadConfig(rwc)
//	if err != nil {
//		return nil, err
//	}
//
//	if len(rootPaths) != 0 {
//		rootKey := strings.Join(rootPaths, ".")
//		newRoot := v.Sub(rootKey)
//
//		// overwrite root iff new root node is valid
//		if newRoot != nil {
//			v = newRoot
//		}
//	}
//
//	s := &Store{
//		Viper: v,
//		rwc:   rwc,
//	}
//
//	return s, nil
//}
//
//func supported(kind string) bool {
//	for _, ext := range viper.SupportedExts {
//		if ext == kind {
//			return true
//		}
//	}
//
//	return false
//}

func init() {
	viper.SupportedExts = append(viper.SupportedExts, "db")
}
