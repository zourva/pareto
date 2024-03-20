package config

import (
	"errors"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"strings"
)

const (
	dbExt = "db"
)

// DBCodec is a customized viper encoder/decoder backed by boltdb.
type DBCodec struct{}

func (DBCodec) Encode(v map[string]any) ([]byte, error) {
	//return yaml.Marshal(v)
	panic("not supported")
	return nil, nil
}

func (DBCodec) Decode(b []byte, v map[string]any) error {
	//return yaml.Unmarshal(b, &v)
	panic("not supported")
	return nil
}

type Flusher = func(map[string]any) error

// Store wraps viper and provides extended functionalities.
// Store uses a storage system as the backlog and accepts
// multiple config sources to merge them into the storage.
//
// A config store is structured as a tree and be flattened
// using a key-value pattern, and the config value is accessed
// using a key, depicted by a tree-node-path.
type Store struct {
	//*viper.Viper
	*koanf.Koanf

	flushers map[string]Flusher
}

type Option = func(s *Store)

func (s *Store) decideProviderParser(f string, t Type) (koanf.Provider, koanf.Parser, error) {
	switch t {
	//case Sqlite:
	//return boltdbProvider(f), boltdbParser(), nil
	//case Boltdb:
	//	return boltdbProvider(f), boltdbParser(), nil
	case Yaml:
		return file.Provider(f), yaml.Parser(), nil
	case Json:
		return file.Provider(f), json.Parser(), nil
	default:
		return nil, nil, errors.New("not supported")
	}
}

func (s *Store) decideTag(t Type) string {
	switch t {
	case Sqlite:
		return Sqlite
	case Boltdb:
		return Boltdb
	case Yaml:
		return Yaml
	case Json:
		return Json
	default:
		return "koanf"
	}
}

// Load loads config from file into this store.
// When called multiple times over different files, configs are merged.
func (s *Store) Load(file string, kind Type, rootKeys ...string) error {
	//tag := s.decideTag(kind)
	root := strings.Join(rootKeys, ".")

	provider, parser, err := s.decideProviderParser(file, kind)
	if err != nil {
		log.Errorf("config type %s invalid: %v", kind, err)
		return err
	}

	// load
	k := koanf.New(".")
	err = k.Load(provider, parser)
	if err != nil {
		log.Errorf("read config file %s failed: %v", file, err)
		return err
	}

	// cut to the root
	if len(root) != 0 {
		k = k.Cut(root)
	}

	// merge
	err = s.Merge(k)
	if err != nil {
		log.Errorf("merge config from %s failed: %v", file, err)
		return err
	}

	log.Infof("config loaded")
	return nil
}

// Flush writes configurations in store back to the given file.
func (s *Store) Flush(key string) error {
	if fn, ok := s.flushers[key]; ok {
		return fn(s.Cut(key).All())
	}

	return errors.New("flusher not found")
}

func (s *Store) MergeStore(store *Store) error {
	return s.Merge(store.Koanf)
}

// UnmarshalKey is here for compatible with viper api.
func (s *Store) UnmarshalKey(path string, o any) error {
	return s.UnmarshalWithConf(path, o, koanf.UnmarshalConf{})
}

// SetDefault is here for compatible with viper api.
func (s *Store) SetDefault(k string, v any) {
	_ = s.Set(k, v)
}

// the default global instance
var store *Store

func init() {
	store = New()
}

func WithFlusher(subPath string, flusher Flusher) Option {
	return func(s *Store) {
		if len(subPath) != 0 && flusher != nil {
			s.flushers[subPath] = flusher
		}
	}
}

// New creates a configuration store.
// Returns the created store or nil if any error occurred.
func New(opts ...Option) *Store {
	s := new(Store)
	s.Koanf = koanf.NewWithConf(koanf.Conf{Delim: "."})
	s.flushers = make(map[string]Flusher)

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
//func Load(file string, rootKeys ...string) error {
//	return store.Load(file, rootKeys...)
//}

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
func Clamp[T box.Number](v *Store, key string, f Getter[T], min, max T) {
	if min > max {
		return
	}

	val := f(key)
	box.Clamp(&val, min, max)
	_ = v.Set(key, val)
}

// ClampDefault acts the same as Clamp except that value of key is overwritten
// by the default value other than the boundary.
func ClampDefault[T box.Number](v *Store, key string, f Getter[T], min, max, def T) {
	if min > max {
		return
	}

	val := f(key)
	box.ClampDefault(&val, min, max, def)
	_ = v.Set(key, val)
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return store.String(key) }

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool { return store.Bool(key) }

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return store.Int(key) }

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string) int32 { return int32(store.Int64(key)) }

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 { return store.Int64(key) }

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string) uint { return uint(store.Int64(key)) }

// GetUint16 returns the value associated with the key as an unsigned integer.
func GetUint16(key string) uint16 { return uint16(store.Int64(key)) }

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string) uint32 { return uint32(store.Int64(key)) }

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 { return uint64(store.Int64(key)) }

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 { return store.Float64(key) }

// Deprecated. use New and Load instead.
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
//func NewStore(file string, rootKeys ...string) (*Store, error) {
//	ok, err := box.PathExists(file)
//	if err != nil {
//		log.Errorf("access file %s failed: %v", file, err)
//		return nil, err
//	}
//
//	if !ok {
//		log.Errorf("file %s does not exist", file)
//		return nil, os.ErrNotExist
//	}
//
//	v := viper.NewWithOptions(
//		viper.KeyDelimiter("."),
//	)
//
//	_ = v.EncoderRegistry().RegisterEncoder(dbExt, DBCodec{})
//	_ = v.DecoderRegistry().RegisterDecoder(dbExt, DBCodec{})
//	v.SetConfigFile(file)
//
//	err = v.ReadInConfig()
//	if err != nil {
//		log.Errorf("read config file %s failed: %v", file, err)
//		return nil, err
//	}
//
//	log.Infof("config loaded")
//
//	root := ""
//	if len(rootKeys) != 0 {
//		root = strings.Join(rootKeys, ".")
//		newRoot := v.Sub(root)
//
//		// overwrite root iff new root node is valid
//		if newRoot != nil {
//			v = newRoot
//		}
//	}
//
//	s := &Store{
//		Viper: v,
//		root:  root,
//	}
//
//	return s, nil
//}
