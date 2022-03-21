package logger

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

// Options defines creation option of logger.
type Options struct {
	Verbosity   string //"v", "vv", or "vvv"
	LogFileName string //stdout/stderr or filename
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
			MaxSize:    50,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		})
	}
}
