package config

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"os"
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
