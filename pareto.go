package pareto

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box/env"
	"github.com/zourva/pareto/box/prof"
	"github.com/zourva/pareto/config"
	"github.com/zourva/pareto/logger"
	"os"
)

// Pareto defines the context.
type Pareto struct {
	layout *env.WorkingDir
	config *config.Store
	logger *logger.Logger
	// diagnoser  *diagnoser.Diagnoser
	// monitor    *monitor.SysMonitor
	// updater    *updater.OtaManager
	profiler  *prof.Profiler
	flagParse bool
}

var pareto *Pareto

func init() {
	pareto = New()
}

// New create a pareto env.
func New() *Pareto {
	p := new(Pareto)

	return p
}

// Config returns the default global instance of config object.
func Config() *config.Store { return pareto.Config() }

func (p *Pareto) Config() *config.Store { return p.config }

// Logger returns the default global instance of logger object.
func Logger() *logger.Logger { return pareto.Logger() }

func (p *Pareto) Logger() *logger.Logger { return p.logger }

// Option defines pareto initialization options.
type Option func(*Pareto)

// EnableFlagParse enables or disables flag.Parse.
func EnableFlagParse(parse bool) Option {
	return func(p *Pareto) {
		p.flagParse = parse
		if parse {
			flag.Parse()
		}
	}
}

// EnableProfiler enables prof.Profiler.
func EnableProfiler() Option {
	return func(p *Pareto) {
		p.profiler = prof.NewProfiler(nil)
		p.profiler.Start()
	}
}

// WithLogger allows to provide a logger config
// as an option.
func WithLogger(l *logger.Logger) Option {
	return func(p *Pareto) {
		p.logger = l
	}
}

// WithLoggerProvider allows to provide a logger create function
func WithLoggerProvider(provider func() *logger.Logger) Option {
	return func(p *Pareto) {
		l := provider()
		if l == nil {
			log.Fatalln("call user provided create logger function failed")
		}
		p.logger = l
	}
}

// WithWorkingDirLayout allows to hint working dir layout.
func WithWorkingDirLayout(wd *env.WorkingDir) Option {
	return func(p *Pareto) {
		p.layout = wd
	}
}

// WithWorkingDir acts the same as WithWorkingDirLayout except that
// it also set system level working dir, using os.Chdir, to the parent
// directory of this executable.
func WithWorkingDir(wd *env.WorkingDir) Option {
	return func(p *Pareto) {
		p.layout = wd
		err := os.Chdir(env.GetExecFilePath() + "/../")
		if err != nil {
			log.Fatalln("change working dir failed:", err)
		}
	}
}

// WithConfigStoreFile specifies a config file to load, which will
// overwrite the default pareto config store.
func WithConfigStoreFile(file string, rootPaths ...string) Option {
	return func(p *Pareto) {
		s, err := config.NewStore(file, rootPaths...)
		if err != nil {
			log.Fatalf("config store load failed: %v", err)
		}

		p.config = s
	}
}

// WithJsonConfParser
// To load the specified configuration file
// and invoke the normalizer function for parsing
func WithJsonConfParser(file string, obj any, normalize func(obj any) error) Option {
	return func(p *Pareto) {
		err := config.LoadJsonConfig(file, obj)
		if err != nil {
			log.Fatalln("load config file(", file, ") failed:", err)
		}

		if normalize != nil {
			err = normalize(obj)
			if err != nil {
				log.Fatalln("call user provided config parse function failed", err)
			}
		}
	}
}

// SetupWithOpts creates a pareto environment for an
// app with the given options.
func SetupWithOpts(options ...Option) {
	for _, fn := range options {
		fn(pareto)
	}

	log.Infoln("setup pareto environment done")
}

// Teardown tears down the working space.
func Teardown() {
	if pareto.profiler != nil {
		pareto.profiler.Stop()
	}

	log.Infoln("teardown pareto environment done")
}
