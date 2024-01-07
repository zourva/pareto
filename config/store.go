package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zourva/pareto/box"
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
	root string //root node path
}

// Load loads config from file into this store.
// When called multiple times over different files, configs are merged.
func (s *Store) Load(file string, rootKeys ...string) error {
	s.SetConfigFile(file)

	err := s.ReadInConfig()
	if err != nil {
		log.Errorf("read config file %s failed: %v", file, err)
		return err
	}

	if len(rootKeys) != 0 {
		s.root = strings.Join(rootKeys, ".")
		newRoot := s.Sub(s.root)

		// overwrite root iff new root node is valid
		if newRoot != nil {
			s.Viper = newRoot
		}
	}

	log.Infof("config loaded")
	return nil
}

// the default global instance
var store *Store

func init() {
	viper.SupportedExts = append(viper.SupportedExts, "db")
	store = New()
}

// New creates a configuration store.
// Returns the created store or nil if any error occurred.
func New() *Store {
	s := new(Store)
	v := viper.NewWithOptions(
		viper.KeyDelimiter("."),
	)

	v.EncoderRegistry().RegisterEncoder("db", DBCodec{})
	v.DecoderRegistry().RegisterDecoder("db", DBCodec{})

	s.Viper = v
	s.root = ""

	return s
}

// GetStore returns the global store instance.
func GetStore() *Store {
	return store
}

// Load loads configurations into the default store,
// with an optional path identifying the node as root
// of the config tree loaded.
//
// The supported config file content has the same set
// of extensions provided by viper (either one of json
// /toml/yaml/yml/properties/props/prop/hcl/tfvars/dotenv
// /env/ini or extended type db)
//
// rootKeys, if provided, will be joined as a full path
// based on the key delimiter, which then identifies a
// subtree, and will be used as the config tree root.
// Error is returned if the given path does not exist.
func Load(file string, rootKeys ...string) error {
	return store.Load(file, rootKeys...)
}

//func Equal(key string, expected any) bool {
//	actual := store.Get(key)
//	if expected == nil || actual == nil {
//		return expected == actual
//	}
//
//	exp, ok := expected.([]byte)
//	if !ok {
//		return reflect.DeepEqual(expected, actual)
//	}
//
//	act, ok := actual.([]byte)
//	if !ok {
//		return false
//	}
//	if exp == nil || act == nil {
//		return exp == nil && act == nil
//	}
//	return bytes.Equal(exp, act)
//}

type Getter[T box.Number] func(string) T

// Clamp overwrites value of key to min or max if its
// value is not within range [min, max].
func Clamp[T box.Number](key string, f Getter[T], min, max T) {
	if min > max {
		return
	}

	val := f(key)
	box.Clamp(&val, min, max)
	store.Set(key, val)
}

// ClampDefault acts the same as Clamp except that value of key is overwritten
// by the default value other than the boundary.
func ClampDefault[T box.Number](key string, f Getter[T], min, max, def T) {
	if min > max {
		return
	}

	val := f(key)
	box.ClampDefault(&val, min, max, def)
	store.Set(key, val)
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return store.GetString(key) }

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool { return store.GetBool(key) }

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return store.GetInt(key) }

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string) int32 { return store.GetInt32(key) }

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 { return store.GetInt64(key) }

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string) uint { return store.GetUint(key) }

// GetUint16 returns the value associated with the key as an unsigned integer.
func GetUint16(key string) uint16 { return store.GetUint16(key) }

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string) uint32 { return store.GetUint32(key) }

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 { return store.GetUint64(key) }

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 { return store.GetFloat64(key) }

// NewStore creates a configuration store based on the given config file.
//
// The supported config file content has the same set of extensions provided
// by viper (either one of json/toml/yaml/yml/properties/props/prop/hcl/tfvars
// /dotenv/env/ini or extended type db)
//
// rootKeys, if provided, will be joined as a full path based on the key delimiter,
// which then identifies a subtree, and will be used as the config tree root and
// if the given path does not exist, a new root node is created.
//
// Returns the created store or nil if any error occurred.
func NewStore(file string, rootKeys ...string) (*Store, error) {
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

	log.Infof("config loaded")

	root := ""
	if len(rootKeys) != 0 {
		root = strings.Join(rootKeys, ".")
		newRoot := v.Sub(root)

		// overwrite root iff new root node is valid
		if newRoot != nil {
			v = newRoot
		}
	}

	s := &Store{
		Viper: v,
		root:  root,
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
