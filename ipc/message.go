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

func NewMessager(conf *MessagerConf) *Messager {
	bus := NewBus(conf.BusConf)
	if bus == nil {
		log.Errorln("NewBus failed")
		return nil
	}

	rpc := NewRPC(conf.RpcConf)
	if rpc == nil {
		log.Errorln("NewRPC failed")
		return nil
	}

	m := &Messager{
		Bus: NewBus(conf.BusConf),
		RPC: NewRPC(conf.RpcConf),
	}

	return m
}
