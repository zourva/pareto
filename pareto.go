package pareto

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/env"
	"github.com/zourva/pareto/logger"
	"github.com/zourva/pareto/prof"
)

type paretoKit struct {
	workingDir *env.WorkingDir
	logger     *logger.Logger
	profiler   *prof.Profiler
}

var bot = new(paretoKit)

type Option func()

func WithLogger(l *logger.Logger) Option {
	return func() {
		bot.logger = l
	}
}

func WithWorkingDir(wd *env.WorkingDir) Option {
	return func() {
		bot.workingDir = wd
	}
}

func SetupWithOpts(options ...Option) {
	for _, o := range options {
		o()
	}
}

// Setup creates the default layout
func Setup() {
	bot.logger = logger.NewLogger(&logger.Options{
		Verbosity:   "v",
		LogFileName: env.GetExecFilePath() + "/../log/out.log",
	})

	bot.workingDir = env.NewWorkingDir(true,
		[]*env.DirInfo{
			{"bin", 0755},
			{"etc", 0755},
			{"lib", 0755},
			{"log", 0755},
		})

	log.Infoln("setup working directory")
}

// Teardown tears down the working space
func Teardown() {
	if bot.profiler != nil {
		bot.profiler.Stop()
	}

	log.Infoln("teardown working directory")
}
