package logger

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

// Options defines creation option of logger.
type Options struct {
	//"v", "vv", or "vvv"
	Verbosity string

	//stdout/stderr or filename
	LogFileName string

	//Max size in MB of the log file before it gets rotated. It defaults to 100MB.
	MaxSize int

	//Max number of days to retain old log files based on the timestamp encoded in their filename.
	//
	//It defaults to 7 days.
	MaxAge int

	//Max number of old log files to retain.
	//It defaults to 3 files.
	//
	//Any files older than MaxAge days are deleted, regardless of MaxBackups.
	//
	//If MaxBackups and MaxAge are both 0, no old log files will be deleted.
	MaxBackups int
}

// Logger abstracts pareto logger.
type Logger struct {
	options *Options
}

// NewLogger creates a new logger with the given options.
func NewLogger(opt *Options) *Logger {
	options := opt
	if options == nil {
		options = &Options{
			Verbosity:   "v",
			LogFileName: "stderr",
		}
	}

	l := &Logger{
		options: options,
	}

	l.setup()

	return l
}

func (l *Logger) setup() {
	// set level
	log.SetLevel(log.InfoLevel)

	if l.options.Verbosity == "v" {
		log.SetLevel(log.DebugLevel)
	} else if l.options.Verbosity == "vv" {
		log.SetLevel(log.TraceLevel)
	} else if strings.Contains(l.options.Verbosity, "vvv") {
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	if l.options.LogFileName == "stderr" ||
		l.options.LogFileName == "stdout" {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(&lumberjack.Logger{
			Filename:   l.options.LogFileName,
			MaxSize:    l.options.MaxSize,
			MaxAge:     l.options.MaxAge,
			MaxBackups: l.options.MaxBackups,
			Compress:   true,
		})
	}
}
