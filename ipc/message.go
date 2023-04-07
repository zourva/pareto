package ipc

import log "github.com/sirupsen/logrus"

// Messager is a communication endpoint peer,
// which combines together messaging patterns, namely Bus and RPC.
// It can acts as an RPC server or client,
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
func NewMessager(conf *MessagerConf) *Messager {
	if conf == nil || (conf.BusConf == nil && conf.RpcConf == nil) {
		log.Errorln("conf is nil, cannot create a messager")
		return nil
	}

	var bus Bus
	if conf.BusConf != nil {
		if bus = NewBus(conf.BusConf); bus == nil {
			log.Errorln("NewBus failed")
			return nil
		}

		log.Infof("bus messager endpoint %s created", conf.BusConf.Name)
	} else {
		log.Infoln("bus messager endpoint creation skipped")
	}

	var rpc RPC
	if conf.RpcConf != nil {
		if rpc = NewRPC(conf.RpcConf); rpc == nil {
			log.Errorln("NewRPC failed")
			return nil
		}

		log.Infof("rpc messager channel %s created", conf.RpcConf.Name)
	} else {
		log.Infoln("rpc messager channel creation skipped")
	}

	m := &Messager{
		Bus: bus,
		RPC: rpc,
	}

	log.Infoln("a messager is created")

	return m
}
