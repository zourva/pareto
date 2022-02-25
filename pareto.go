package pareto

import (
	log "github.com/sirupsen/logrus"
	"pareto/env"
	"pareto/logger"
	"pareto/prof"
)

type paretoKit struct {
	workingDir *env.WorkingDir
	logger     *logger.Logger
	profiler   *prof.Profiler
}

var bot = new(paretoKit)

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
