package pareto

import (
	"github.com/zourva/pareto/env"
	"github.com/zourva/pareto/logger"
	"testing"
)

func TestSetupDefault(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "working dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup()
			Teardown()
		})
	}
}

func TestSetupWithOpts(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "customize logger and wd",
			args: args{
				options: []Option{
					WithLogger(
						logger.NewLogger(&logger.Options{
							Verbosity:   "vv",
							LogFileName: env.GetExecFilePath() + "/../log/22.log",
						}),
					),
					WithWorkingDir(
						env.NewWorkingDir(true,
							[]*env.DirInfo{
								{"bin", 0755},
								{"etc", 0755},
								{"lib", 0755},
								{"log", 0755},
								{"data", 0755},
								{"installer", 0755},
							}),
					),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupWithOpts(tt.args.options...)
			Teardown()
		})
	}
}
