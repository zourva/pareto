package box

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

// PathExists returns true if path exist and false otherwise with nil error
//or error and false when error occurred
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	log.Errorln("error when check stat of", path, err)

	return false, err
}

// GetWorkingDir returns the base dir path of the
// executable that is calling this function.
func GetWorkingDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Errorln("error when get working dir", err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

// ProcessExists returns true & nil if exists, or
// false & error when error, or,
// false & nil when not exists
func ProcessExists(pid uint32) (bool, error) {
	_, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}

	path := fmt.Sprintf("/proc/%d", pid)
	_, err = os.Stat(path)
	if err != nil {
		//ok, not exist
		return false, nil
	}

	return true, nil
}
