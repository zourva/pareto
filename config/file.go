package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"os"
	"reflect"
	"time"
)

// LoadJsonConfig loads config object of a JSON file
// and deserialize to the given object.
// Return nil or any error if happened during loading and unmarshal.
func LoadJsonConfig(file string, obj any) error {
	ok, err := box.PathExists(file)
	if err != nil {
		log.Errorf("config file(%s) not available:%v", file, err)
		return err
	}

	if !ok {
		log.Errorf("config file(%s) not exist", file)
		return os.ErrNotExist
	}

	buf, err := os.ReadFile(file)
	if err != nil {
		log.Errorln("load config file failed:", err)
		return err
	}

	err = json.Unmarshal(buf, obj)
	if err != nil {
		log.Errorln("unmarshal config file failed:", err)
		return err
	}

	var dst bytes.Buffer
	_ = json.Compact(&dst, buf)

	log.Traceln("json config loaded: ", dst.String())

	return nil
}

type Kind int

const (
	JSON Kind = iota
	YAML
)

func (k Kind) String() string {
	switch k {
	case JSON:
		return "JSON"
	case YAML:
		return "YAML"
	}
	return "unknown"
}

type File[T any] struct {
	FileName   string
	FileKind   Kind
	ModifyTime time.Time
	Content    T
}

func NewFile[T any]() *File[T] {
	return &File[T]{}
}

func (f *File[T]) Init(file string, kind Kind) (*T, error) {
	modify, err := f.modifyTime(file)
	if err != nil {
		log.Errorf("get file(%s) failed, err:%v", file, err)
		return nil, err
	}

	err = f.load(file, kind, &f.Content)
	if err != nil {
		log.Errorf("load file(%s) failed, err:%v", file, err)
		return nil, err
	}

	f.FileName = file
	f.FileKind = kind
	f.ModifyTime = modify

	log.Infof("read file(%s,%s) successfully", file, kind)
	return &f.Content, nil
}

func (f *File[T]) Listen(ctx context.Context, duration time.Duration, changed func(content *T) error) {
	ticker := time.NewTicker(duration)

	log.Infof("listening file ...")
	for {
		select {
		case <-ctx.Done():
			log.Infof("stop listening, because received context done")
			return
		case <-ticker.C:
			err := f.checker(f.FileName, f.FileKind, changed)
			if nil != err {
				log.Infof("stop listening, because check file failed, err:%v", err)
				return
			}
		}
	}
}

func (f *File[T]) load(file string, kind Kind, content *T) error {
	switch kind {
	case JSON:
		return LoadJsonConfig(file, content)
	case YAML:
		return fmt.Errorf("not supported yaml file")
	}

	return nil
}

func (f *File[T]) modifyTime(file string) (time.Time, error) {
	info, err := os.Stat(file)
	if err != nil {
		log.Errorf("stat file(%s) failed,err:%v", file, err)
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

func (f *File[T]) checker(file string, kind Kind, changed func(content *T) error) error {
	modify, err := f.modifyTime(file)
	if err != nil {
		log.Errorf("get file(%s) failed, err:%v", file, err)
		return nil
	}

	var content T
	err = f.load(file, kind, &content)
	if err != nil {
		log.Errorf("load file(%s) failed, err:%v", file, err)
		return nil
	}

	if reflect.DeepEqual(f.Content, content) {
		return nil
	}

	log.Infof("file(%s) was changed(%v), trigger update...", file, modify.Format(time.RFC3339))

	err = changed(&content)
	if nil != err {
		log.Errorf("call user function failed, %s", err)
		return nil
	}

	f.Content = content
	f.ModifyTime = modify
	log.Infof("file(%s) was changed(%v), trigger update successfully", file, modify.Format(time.RFC3339))
	return nil
}
