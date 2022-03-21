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

// Option defines pareto initialization options.
type Option func()

// WithLogger allows to provide a logger config
// as an option.
func WithLogger(l *logger.Logger) Option {
	return func() {
		bot.logger = l
	}
}

// WithWorkingDir allows to hint working dir layout.
func WithWorkingDir(wd *env.WorkingDir) Option {
	return func() {
		bot.workingDir = wd
	}
}

// SetupWithOpts create a pareto environment with the
// given options.
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
			{Name: "bin", Mode: 0755},
			{Name: "etc", Mode: 0755},
			{Name: "lib", Mode: 0755},
			{Name: "log", Mode: 0755},
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
