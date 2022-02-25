package prof

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"pareto/box"
	"runtime"
	"runtime/pprof"

	_ "net/http/pprof"
)

type Options struct {
	CpuFile *os.File
	MemFile *os.File

	// http export addr, if empty, no http access
	Endpoint string
}

type Profiler struct {
	options *Options
	started bool
}

func NewProfiler(opt *Options) *Profiler {
	o := &Options{}

	if opt != nil {
		o = opt
	} else {
		var err error
		if o.CpuFile, err = os.Create("cpu.prof"); err != nil {
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

func (p *Profiler) Start() {
	p.commit()
	p.started = true
}

func (p *Profiler) Stop() {
	p.started = false

	if p.options.CpuFile != nil {
		_ = p.options.CpuFile.Close()
		pprof.StopCPUProfile()
	}

	if p.options.MemFile != nil {
		// get up-to-date statistics
		runtime.GC()
		_ = pprof.WriteHeapProfile(p.options.MemFile)
		_ = p.options.MemFile.Close()
	}
}

func (p *Profiler) WithCpuFile(f *os.File) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if f != nil {
		p.options.CpuFile = f
	}
}

func (p *Profiler) WithMemFile(f *os.File) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if f != nil {
		p.options.MemFile = f
	}
}

func (p *Profiler) WithHttpEndpoint(ep string) {
	if p.started {
		log.Warnln("profiler is running already, ignored")
		return
	}

	if len(ep) > 0 {
		p.options.Endpoint = ep
	}
}

func (p *Profiler) commit() {
	if p.options.CpuFile != nil {
		f := p.options.CpuFile
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
