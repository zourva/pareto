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
	config *config.Store
	logger *logger.Logger
	// diagnoser  *diagnoser.Diagnoser
	// monitor    *monitor.SysMonitor
	// updater    *updater.OtaManager
	//profiler *prof.Profiler

	disableFlags  bool
	disableLogger bool
	configFile    string
	configRoot    string
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

func (p *Pareto) Config() *config.Store { return p.config }

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
	if len(p.configFile) != 0 {
		if err := cfg.Load(p.configFile, p.configRoot); err != nil {
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
	p.config = config.GetStore()
	err := p.config.MergeConfigMap(cfg.AllSettings())
	if err != nil {
		log.Fatalf("config store merge failed: %v", err)
	}

	// create logger
	if !p.disableLogger {
		if p.loggerCreator != nil {
			l := p.loggerCreator()
			if l == nil {
				log.Fatalln("create logger using provider failed")
			}
			p.logger = l
		} else {
			cfg := logger.Options{}
			if err := p.config.UnmarshalKey("logger", &cfg); err != nil {
				log.Fatalln("create logger failed:", err)
			}
			p.logger = logger.NewLogger(&cfg)
		}
	}
}

// Option defines pareto initialization options.
type Option func(*Pareto)

//// Config defines common configuration framework.
//type Config struct {
//	Service service.Descriptor `json:"service" yaml:"service"`
//	Logger  logger.Options     `json:"logger" yaml:"logger"`
//	App     any                `json:"app" yaml:"app"`
//}

type ConfigNormalizer = func(v *config.Store) error
type ConfigDefaultsProvider = func(v *config.Store)
type LoggerProvider = func() *logger.Logger

func DefaultNormalize(v *config.Store) error {
	// FixMe: viper doesn't support partial override when UnmarshalKey
	// consider using https://github.com/knadh/koanf which
	// overrides by loading order.
	// If we ever used Set() to update any variable,
	// then all variables must be overwritten by Set
	// or else we get empty values for those not overridden.
	// As the underlying map, which is `override map[string]any`, is
	// not merged with `config map[string]any` by default.
	// So, we need to make changes over a temp store and then merge it into
	// the global config instance.
	config.ClampDefault(v, "logger.maxSize", v.GetInt, 20, 100, 50)
	config.ClampDefault(v, "logger.maxAge", v.GetInt, 1, 30, 7)
	config.ClampDefault(v, "logger.maxBackups", v.GetInt, 0, 20, 3)

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
func WithConfigStore(file string, rootKeys ...string) Option {
	return func(p *Pareto) {
		p.configFile = file
		p.configRoot = strings.Join(rootKeys, ".")
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
