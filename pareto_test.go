package pareto

import (
	"github.com/zourva/pareto/box/env"
	"github.com/zourva/pareto/logger"
	"testing"
)

func TestSetupDefault(t *testing.T) {
	Setup()
	Teardown()
}

func TestSetupWithOpts(t *testing.T) {
	options := []Option{
		WithLogger(
			logger.NewLogger(&logger.Options{
				Verbosity:   "vv",
				LogFileName: env.GetExecFilePath() + "/../log/22.log",
			}),
		),
		WithWorkingDir(
			env.NewWorkingDir(true,
				[]*env.DirInfo{
					{Name: "bin", Mode: 0755},
					{Name: "etc", Mode: 0755},
					{Name: "lib", Mode: 0755},
					{Name: "log", Mode: 0755},
					{Name: "data", Mode: 0755},
					{Name: "installer", Mode: 0755},
				}),
		),
	}

	SetupWithOpts(options...)
	Teardown()
}
