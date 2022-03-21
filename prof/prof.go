package prof

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	// profiling
	_ "net/http/pprof"
)

// Options defines profiler functionalities.
type Options struct {
	CPUFile *os.File
	MemFile *os.File

	// http export addr, if empty, no http access
	Endpoint string
}

// Profiler manages the profiling process.
type Profiler struct {
	options *Options
	started bool
}

// NewProfiler creates a new profiler with the given options.
func NewProfiler(opt *Options) *Profiler {
	o := &Options{}

	if opt != nil {
		o = opt
	} else {
		var err error
		if o.CPUFile, err = os.Create("cpu.prof"); err != nil {
			log.Fatalln("start cpu profiling failed", err)
		}

		o.MemFile, err = os.Create("mem.prof")
		if err != nil {
			log.Fatalln("start mem profiling failed", err)
		}

		o.Endpoint = "localhost:6060"
	}

	p := &Profiler{
		options: o,
	}

	return p
}

// Start starts the profiler.
func (p *Profiler) Start() {
	p.commit()
	p.started = true
}

// Stop stops the profiler.
func (p *Profiler) Stop() {
	p.started = false

	if p.options.CPUFile != nil {
		_ = p.options.CPUFile.Close()
		pprof.StopCPUProfile()
	}

	if p.options.MemFile != nil {
		// get up-to-date statistics
		runtime.GC()
		_ = pprof.WriteHeapProfile(p.options.MemFile)
		_ = p.options.MemFile.Close()
	}
}

// WithCPUFile sets the cpu profiling file.
// The old one, if any, will be closed.
func (p *Profiler) WithCPUFile(f *os.File) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if p.options.CPUFile != nil {
		_ = p.options.CPUFile.Close()
	}

	if f != nil {
		p.options.CPUFile = f
	}
}

// WithMemFile sets the memory profiling file.
// The old one, if any, will be closed.
func (p *Profiler) WithMemFile(f *os.File) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if p.options.MemFile != nil {
		_ = p.options.MemFile.Close()
	}

	if f != nil {
		p.options.MemFile = f
	}
}

// WithHTTPEndpoint sets the endpoint, which is set to
// localhost:6060 by default.
func (p *Profiler) WithHTTPEndpoint(ep string) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if len(ep) > 0 {
		p.options.Endpoint = ep
	}
}

func (p *Profiler) commit() {
	if p.options.CPUFile != nil {
		f := p.options.CPUFile
		_ = f.Close()
		pprof.StopCPUProfile()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalln("start cpu profiling failed", err)
		}

		log.Infoln("enable cpu profiling")
	}

	if p.options.MemFile != nil {
		// get up-to-date statistics
		runtime.GC()
		//_ = pprof.WriteHeapProfile(p.options.MemFile)
		//_ = p.options.MemFile.Close()

		//f := p.options.MemFile

		//_ = pprof.WriteHeapProfile(f)

		log.Infoln("enable mem profiling")
	}

	if box.ValidateEndpoint(p.options.Endpoint) {
		go func() {
			log.Fatalln(http.ListenAndServe(p.options.Endpoint, nil))
		}()
	}
}
