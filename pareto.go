package pareto

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box/env"
	"github.com/zourva/pareto/config"
	"github.com/zourva/pareto/logger"
	"os"
	"strings"
)

// Pareto defines the context.
type Pareto struct {
	layout *env.WorkingDir
	logger *logger.Logger
	//config *config.Store
	// diagnoser  *diagnoser.Diagnoser
	// monitor    *monitor.SysMonitor
	// updater    *updater.OtaManager
	//profiler *prof.Profiler

	disableFlags  bool
	disableLogger bool

	confFile string //root config file
	confData string //root config path

	confFileType config.FileType // root config file type
	confDataType config.DataType //root config data type

	defaults      ConfigDefaultsProvider
	normalize     ConfigNormalizer
	loggerCreator LoggerProvider
}

var pareto *Pareto

func init() {
	pareto = New()
}

// New create a pareto env.
func New() *Pareto {
	p := new(Pareto)
	p.disableFlags = false
	p.disableLogger = false
	p.defaults = DefaultConfig
	p.normalize = DefaultNormalize

	return p
}

// Config returns the default global instance of config object.
func Config() *config.Store { return pareto.Config() }

func (p *Pareto) Config() *config.Store { return config.GetStore() }

// Logger returns the default global instance of logger object.
func Logger() *logger.Logger { return pareto.Logger() }

func (p *Pareto) Logger() *logger.Logger { return p.logger }

func (p *Pareto) Setup() {
	if !p.disableFlags {
		flag.Parse()
	}

	// change working dir
	if err := os.Chdir(env.GetExecFilePath() + "/../"); err != nil {
		log.Fatalln("change working dir failed:", err)
	}

	// create app config which will be merged
	cfg := config.New()

	// set defaults
	p.defaults(cfg)

	// load config
	if len(p.confFile) != 0 {
		if err := cfg.Load(p.confFile, p.confFileType, p.confDataType, p.confData); err != nil {
			log.Fatalf("config store load failed: %v", err)
		}
	}

	// normalize
	if p.normalize != nil {
		if err := p.normalize(cfg); err != nil {
			log.Fatalf("config store normalize failed: %v", err)
		}
	}

	// merge with the global
	err := config.MergeStore(cfg)
	if err != nil {
		log.Fatalf("config store merge failed: %v", err)
	}

	// create logger using global config
	if !p.disableLogger {
		if p.loggerCreator != nil {
			l := p.loggerCreator()
			if l == nil {
				log.Fatalln("create logger using provider failed")
			}
			p.logger = l
		} else {
			options := logger.Options{}
			if e := config.UnmarshalKey("logger", &options); e != nil {
				log.Fatalln("create logger failed:", e)
			}
			p.logger = logger.NewLogger(&options)
		}
	}
}

// Option defines pareto initialization options.
type Option func(*Pareto)

type ConfigNormalizer = func(v *config.Store) error
type ConfigDefaultsProvider = func(v *config.Store)
type LoggerProvider = func() *logger.Logger

func DefaultNormalize(v *config.Store) error {
	config.ClampDefault(v, "logger.maxSize", v.Int, 20, 100, 50)
	config.ClampDefault(v, "logger.maxAge", v.Int, 1, 30, 7)
	config.ClampDefault(v, "logger.maxBackups", v.Int, 0, 20, 3)

	return nil
}

func DefaultConfig(v *config.Store) {
	v.SetDefault("service.name", "pareto")
	v.SetDefault("service.registry", "nats://127.0.0.1:4222")
	v.SetDefault("logger.verbosity", "vv")
	v.SetDefault("logger.logFileName", "stdout")
	v.SetDefault("logger.maxBackups", 3)
	v.SetDefault("logger.maxSize", 50)
	v.SetDefault("logger.maxAge", 7)
}

// DisableFlags disables flag.Parse.
func DisableFlags() Option {
	return func(p *Pareto) {
		p.disableFlags = true
	}
}

func DisableLogger() Option {
	return func(p *Pareto) {
		p.disableLogger = true
	}
}

// WithLogger allows to provide a logger config
// as an option.
func WithLogger(l *logger.Logger) Option {
	return func(p *Pareto) {
		p.logger = l
	}
}

// WithLoggerProvider allows to provide a logger create function.
func WithLoggerProvider(provider func() *logger.Logger) Option {
	return func(p *Pareto) {
		p.loggerCreator = provider
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
	}
}

// WithConfigStore specifies a config file to load, which will
// overwrite the default pareto config store.
func WithConfigStore(file string, ft config.FileType, dt config.DataType, rootKeys ...string) Option {
	return func(p *Pareto) {
		p.confFile = file
		p.confDataType = dt
		p.confFileType = ft
		p.confData = strings.Join(rootKeys, ".")
	}
}

func WithConfigDefaultsProvider(fn ConfigDefaultsProvider) Option {
	return func(p *Pareto) {
		p.defaults = fn
	}
}

// WithConfigNormalizer provides a config normalizer function
// which will be called when the config, if exists, is loaded.
func WithConfigNormalizer(fn ConfigNormalizer) Option {
	return func(p *Pareto) {
		p.normalize = fn
	}
}

// WithJsonConfParser loads the json config file and invokes the normalize function.
//
// Deprecated: WithJsonConfParser accepts json format config file only, and is outdated.
// Use WithConfigDefaultsProvider, WithConfigNormalizer and WithConfigStore instead.
func WithJsonConfParser(file string, obj any, normalize func(obj any) error) Option {
	return func(p *Pareto) {
		err := config.LoadJsonConfig(file, obj)
		if err != nil {
			log.Fatalf("load config file %s failed: %v", file, err)
		}

		if normalize != nil {
			err = normalize(obj)
			if err != nil {
				log.Fatalln("call normalize function failed:", err)
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

	pareto.Setup()

	log.Infoln("setup pareto environment done")
}

// Teardown tears down the working space.
func Teardown() {
	log.Infoln("teardown pareto environment done")
}
