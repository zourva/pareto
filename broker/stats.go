package broker

import (
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	"time"
)

type counters struct {
	bytesSent uint64 //number bytes sent
	bytesRecv uint64 //number bytes received
	requests  uint64 //number requests sent
	replies   uint64 //number replies received
}

func (c *counters) clear() {
	c.bytesSent = 0
	c.bytesRecv = 0
	c.requests = 0
	c.replies = 0
}

type kpis struct {
	rps    uint64
	ulRate uint64
	dlRate uint64
	delay  int64
}

func (k *kpis) clear() {
	k.rps = 0
	k.ulRate = 0
	k.dlRate = 0
	k.delay = 0
}

type statSession struct {
	startTime int64
	stopTime  int64
	counters  counters
}

func (s *statSession) hackStart(szSent, nrReq uint64) {
	s.startTime = time.Now().Unix()
	s.counters.requests += nrReq
	s.counters.bytesSent += szSent
}

func (s *statSession) hackStop(szRecv, nrRep uint64) {
	s.stopTime = time.Now().Unix()
	s.counters.bytesRecv += szRecv
	s.counters.replies += nrRep
}

type statistician struct {
	sampleTime int64
	counters   counters
	kpis       kpis
}

func newStatistician() *statistician {
	s := &statistician{
		sampleTime: time.Now().Unix(),
	}

	return s
}

func (s *statistician) session() *statSession {
	return &statSession{}
}

func (s *statistician) updateCounters(ssn *statSession) {
	s.kpis.delay = box.MaxI64(s.kpis.delay, ssn.stopTime-ssn.startTime)
	s.counters.bytesSent += ssn.counters.bytesSent
	s.counters.bytesRecv += ssn.counters.bytesRecv
	s.counters.requests += ssn.counters.requests
	s.counters.replies += ssn.counters.replies
}

func (s *statistician) sample(reset bool) *kpis {
	now := time.Now().Unix()
	duration := now - s.sampleTime
	log.Traceln("now, start, duration", now, s.sampleTime, duration)
	if duration == 0 {
		log.Traceln("sample time too short, skip this round")
		return &s.kpis
	}

	s.kpis.rps = s.counters.requests / uint64(duration)
	s.kpis.dlRate = s.counters.bytesRecv / uint64(duration)
	s.kpis.ulRate = s.counters.bytesSent / uint64(duration)

	if reset {
		s.kpis.clear()
		s.counters.clear()
	}

	s.sampleTime = now

	return &s.kpis
}
