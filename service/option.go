package service

import "github.com/zourva/pareto/ipc"

type Option func(s *MetaService)

func WithMessager(m *ipc.Messager) Option {
	return func(s *MetaService) {
		s.messager = m
	}
}

func WithRegistrar(reg *Registrar) Option {
	return func(s *MetaService) {
		s.registrar = reg
	}
}

func EnableTrace(on bool) Option {
	return func(s *MetaService) {
		s.enableTrace = on
	}
}

func WithStatusConfig(c *StatusConf) Option {
	return func(s *MetaService) {
		s.conf = c
	}
}
