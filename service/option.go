package service

import "github.com/zourva/pareto/ipc"

type Option func(s *MetaService)

func WithMessager(m *ipc.Messager) Option {
	return func(s *MetaService) {
		s.messager = m
	}
}

func WithRegistrar(reg Registrar) Option {
	return func(s *MetaService) {
		s.registrar = reg
	}
}

func EnableTrace(on bool) Option {
	return func(s *MetaService) {
		s.enableTrace = on
	}
}

// WithPrivateChannelHandler provides an RR handler for endpoint
// bound on /registry-center/service/handle/{service-name}.
func WithPrivateChannelHandler(handler ipc.CalleeHandler) Option {
	return func(s *MetaService) {
		s.handler = handler
	}
}

func WithStatusConfig(c *StatusConf) Option {
	return func(s *MetaService) {
		s.conf = c
	}
}
