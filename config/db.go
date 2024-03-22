package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"time"
)

// storage abstracts the config
// store persistent layer.
type storage struct {
	//options
	path    string
	options *bolt.Options

	//instance
	inst *bolt.DB
}

// newStorage creates a new configuration
// store using the given file name.
func newStorage(file string) *storage {
	db := &storage{
		path:    file,
		options: &bolt.Options{Timeout: 5 * time.Second},
	}

	return db
}

// open creates and opens a database at the given path.
// If the file does not exist it will be created automatically.
func (d *storage) open() error {
	db, err := bolt.Open(d.path, 0644, d.options)
	if err != nil {
		log.Errorf("open storage file %s failed %v", d.path, err)
		return err
	}

	d.inst = db

	return nil
}

func (d *storage) close() error {
	if d.inst == nil {
		return nil
	}

	return d.inst.Close()
}

func (d *storage) upsert(table string, key string, value []byte) error {
	err := d.inst.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte(table))
		if e != nil {
			return fmt.Errorf("create table failed: %s", e)
		}

		return b.Put([]byte(key), value)
	})

	return err
}

func (d *storage) queryOne(table string, key string) ([]byte, error) {
	var value []byte = nil
	err := d.inst.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not found", table)
		}

		value = b.Get([]byte(key))
		return nil
	})

	return value, err
}

func (d *storage) queryAll(table string, iterator func(k, v []byte) error) error {
	err := d.inst.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not found", table)
		}

		return b.ForEach(iterator)
	})

	return err
}

func (d *storage) deleteOne(table string, key string) error {
	err := d.inst.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not exist", table)
		}

		return b.Delete([]byte(key))
	})

	return err
}
