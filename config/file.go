package config

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"os"
)

func LoadJsonConfig(file string, obj any) error {
	if ok, err := box.PathExists(file); err != nil || !ok {
		log.Errorf("config file(%s) not available:%v", file, err)
		return err
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
	json.Indent(&dst, buf, "\n", "")
	log.Infoln("config file loaded: ", string(buf))
	return nil
}
