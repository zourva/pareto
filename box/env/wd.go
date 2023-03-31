package env

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

// DirInfo describes directory entry info.
type DirInfo struct {
	Name string
	Mode os.FileMode
}

// WorkingDir defines the working directory of an app.
type WorkingDir struct {
	// Working directory of this distribution
	Path string

	// Path of the executable with file name excluded
	ExecPath string

	// Sub-directories in the working dir
	Directories []string
}

// NewWorkingDir creates a working directory layout
// according to the given directory mappings.
//
// Working dir is derived from the executable file, and if parent is true,
// it is the parent dir of the exe file, otherwise it's the path where the exe file located.
//
// Directories will be created with the given names and/or paths in dirs.
// For example, to create a working dir layout of
//
//	   demo/
//		    ├─bin/
//	     │  └─test.exe
//	     ├─data/
//	     ├─etc/
//	     │  └─conf.db
//	     ├─lib/
//	     └─log/
//	         └─error.log
//
//	 call
//	 NewWorkingDir(true, []DirInfo{
//			{Name: "bin", Mode: 0755},
//			{Name: "data", Mode: 0755},
//			{Name: "etc", Mode: 0755},
//			{Name: "lib", Mode: 0755},
//			{Name: "log", Mode: 0755},
//	  }]
//
// NOTE: This method is idempotent.
func NewWorkingDir(parent bool, dirs []*DirInfo) *WorkingDir {
	wd := &WorkingDir{
		Directories: []string{},
	}

	wd.normalizeWorkingDir(parent, dirs)

	return wd
}

// normalizeWorkingDir normalize a working directory layout
// according to the given directory mappings.
//
// NOTE: This method is idempotent.
func (dir *WorkingDir) normalizeWorkingDir(parent bool, dirs []*DirInfo) {
	path, err := os.Executable()
	if err != nil {
		log.Fatalln("get working dir failed", err)
	}

	exe, err := filepath.EvalSymlinks(path)
	if err != nil {
		log.Fatalln("get working dir failed", err)
	}

	execPath := filepath.Dir(exe)
	wd := execPath
	if parent {
		wd = filepath.Dir(execPath + "/../")
	}

	dir.Path = wd
	dir.ExecPath = execPath

	for _, d := range dirs {
		p := filepath.Dir(wd + "/" + d.Name + "/")
		err := os.MkdirAll(p, d.Mode)
		if err != nil {
			log.Fatalf("create dir %s failed: %v", d.Name, err)
		}

		dir.Directories = append(dir.Directories, p)
	}
}

// GetPath returns the working dir full path.
func (dir *WorkingDir) GetPath() string {
	return dir.Path
}

// GetExecFilePath returns the directory part of the full path of exec.
//
//	 e.g.: when running file.exe, located at
//			/path/to/some/file.exe,
//	 this function will return
//			/path/to/some/
func GetExecFilePath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Errorln("error when get working dir", err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}
