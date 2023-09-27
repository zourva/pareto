package ipc

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

// Messager is a communication endpoint peer,
// which combines messaging patterns, namely Bus and RPC.
// It acts as an RPC server or client,
// a BUS subscriber or BUS publisher, or both.
type Messager struct {
	Bus
	RPC
}

type MessagerConf struct {
	BusConf *BusConf
	RpcConf *RPCConf
}

// NewMessager creates a messager using the given config.
//
//	NOTE:
//	Bus endpoint will be created iff BusConf is not nil.
//	RPC channel will be created iff RPCConf is not nil.
//
// Returns nil when both conf are nil.
func NewMessager(conf *MessagerConf) (*Messager, error) {
	if conf == nil || (conf.BusConf == nil && conf.RpcConf == nil) {
		log.Errorln("conf is nil, cannot create a messager")
		return nil, errors.New("messager config is invalid")
	}

	var bus Bus
	var err error
	if conf.BusConf != nil {
		if bus, err = NewBus(conf.BusConf); err != nil {
			log.Errorln("NewBus failed")
			return nil, err
		}

		log.Infof("type %d bus endpoint %s(%p) created",
			conf.BusConf.Type, conf.BusConf.Name, bus)
	} else {
		log.Infoln("bus endpoint creation skipped")
	}

	var rpc RPC
	if conf.RpcConf != nil {
		if rpc, err = NewRPC(conf.RpcConf); err != nil {
			log.Errorln("NewRPC failed")
			return nil, err
		}

		log.Infof("type %d rpc channel %s(%p) created",
			conf.RpcConf.Type, conf.RpcConf.Name, rpc)
	} else {
		log.Infoln("rpc channel creation skipped")
	}

	m := &Messager{
		Bus: bus,
		RPC: rpc,
	}

	log.Infof("messager %p is created", m)

	return m, nil
}
