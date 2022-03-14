package node

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zourva/pareto/box"
	bolt "go.etcd.io/bbolt"
	"os"
	"time"
)

// persistent layer
type confManager struct {
	loaded  bool
	path    string
	inst    *bolt.DB
	options *bolt.Options
}

const (
	confTable = "agentConf"
	nodeTable = "agentNodes"
	confKey   = "agentConf"
)

// Agent config, static & semi-static & dynamic
type Conf struct {
	Identity       string `json:"identity"`       //identity of current agent
	ExpirationTime int64  `json:"expirationTime"` //current provision expiration time
	ProvisionTime  int64  `json:"provisionTime"`  //initial provision timestamp
	LastUpdateTime int64  `json:"lastUpdateTime"` //timestamp when recent update of provision info
	RepeatTimes    uint32 `json:"repeatTimes"`    // number of times re-provisioned
}

type Node struct {
	Identity   string `json:"identity"`   //identity of a node
	Status     uint32 `json:"status"`     //node status
	Endpoint   string `json:"endpoint"`   //current endpoint address
	ExpireTime uint64 `json:"expireTime"` //identity expiration time, in seconds
	SignUpTime uint64 `json:"signUpTime"` //sign up timestamp, in seconds
	SignInTime uint64 `json:"signInTime"` //sign in timestamp, in seconds
	UpdateTime uint64 `json:"updateTime"` //timestamp when recent update of provision info
}

type AgentConfManager interface {
	GetConf() *Conf
	SaveConf(c *Conf) error
	NeedProvision() bool
	UpdateProvision(newId string, expire uint64) bool
}

type ServerConfManager interface {
	GetNode(id string) *Node
	SaveNode(n *Node) error
}

// NewAgentConfManager creates a AgentConfManager
// using the given path as the conf.db file
func NewAgentConfManager(path string) AgentConfManager {
	cm := newConfManager(path)

	if cm.GetConf() == nil {
		// write the default data when it's the first time
		if cm.SaveConf(defaultConf()) != nil {
			log.Errorln("init conf failed")
			return nil
		}

		log.Infoln("conf initialized")
	}

	return cm
}

// NewServerConfManager creates a ServerConfManager
// using the given path as the conf.db file
func NewServerConfManager(path string) ServerConfManager {
	return newConfManager(path)
}

func newConfManager(path string) *confManager {
	conf := &confManager{
		loaded:  false,
		path:    path,
		options: &bolt.Options{Timeout: 10 * time.Second},
		inst:    nil,
	}

	if !conf.load() {
		log.Errorln("load conf failed")
		return nil
	}

	return conf
}

func defaultConf() *Conf {
	return &Conf{
		Identity:       "",
		ExpirationTime: 0,
		ProvisionTime:  0,
		LastUpdateTime: 0,
		RepeatTimes:    0,
	}
}

func (s *confManager) NeedProvision() bool {
	c := s.GetConf()
	id := c.Identity

	if len(id) == 0 {
		return true
	}

	if time.Now().Unix() >= c.ExpirationTime {
		log.Infoln("identity expired, re-provisioning is needed")
		return true
	}

	return false
}

func (s *confManager) UpdateProvision(newId string, expire uint64) bool {
	conf := s.GetConf()
	conf.Identity = newId
	conf.ExpirationTime = int64(expire)
	conf.LastUpdateTime = time.Now().Unix()
	conf.RepeatTimes += 1

	// update only when it's written 1st time
	if conf.ProvisionTime == 0 {
		conf.ProvisionTime = conf.LastUpdateTime
	}

	return s.SaveConf(conf) == nil
}

func (s *confManager) GetConf() *Conf {
	b, e := s.queryOne(confTable, confKey)
	if e != nil {
		return nil
	}

	if b == nil {
		log.Infoln("conf not found:", confKey)
		return nil
	}

	c := &Conf{}
	if e = json.Unmarshal(b, c); e != nil {
		log.Errorln("unmarshal failed", e)
		return nil
	}

	return c
}

func (s *confManager) SaveConf(c *Conf) error {
	b, _ := json.Marshal(c)

	return s.upsert(confTable, confKey, b)
}

func (s *confManager) GetNode(id string) *Node {
	b, e := s.queryOne(nodeTable, id)
	if e != nil {
		return nil
	}

	if b == nil {
		log.Infoln("no node found for:", id)
		return nil
	}

	n := &Node{}
	if e = json.Unmarshal(b, n); e != nil {
		log.Errorln("unmarshal failed", e)
		return nil
	}

	return n
}

func (s *confManager) SaveNode(n *Node) error {
	b, _ := json.Marshal(n)

	return s.upsert(nodeTable, n.Identity, b)
}

// Load loads conf from config db file
func (s *confManager) load() bool {
	_, err := box.PathExists(s.path)
	if err != nil {
		log.Errorln("check path existence failed")
		return false
	}

	if !s.open() {
		log.Errorln("open conf file failed")
		return false
	}

	//if ok {
	//	//just loaded
	//	log.Infoln("conf loaded")
	//} else {
	//	// write the default data when it's the first time
	//	if s.SaveConf(defaultConf()) != nil {
	//		log.Errorln("init conf failed")
	//		return false
	//	}
	//
	//	log.Infoln("conf initialized")
	//}
	//just loaded
	log.Infoln("conf loaded")

	s.loaded = true

	return true
}

// Reset close and delete the db file
func (s *confManager) reset() error {
	if err := s.close(); err != nil {
		log.Errorln("close db error", err)
		return err
	}

	if err := os.Remove(s.path); err != nil {
		log.Errorln("remove db file error", err)
		return err
	}

	log.Infoln("factory reset done")

	return nil
}

// open creates and opens a database at the given path.
// If the file does not exist then it will be created automatically.
func (s *confManager) open() bool {
	dd, err := bolt.Open(s.path, 0644, s.options)
	if err != nil {
		log.Errorln("open db failed:", err)
		return false
	}

	s.inst = dd

	return true
}

func (s *confManager) close() error {
	if s.inst == nil {
		return nil
	}

	s.loaded = false

	return s.inst.Close()
}

func (s *confManager) upsert(table string, key string, value []byte) error {
	err := s.inst.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte(table))
		if e != nil {
			return fmt.Errorf("create table failed: %s", e)
		}

		return b.Put([]byte(key), value)
	})

	return err
}

func (s *confManager) queryOne(table string, key string) ([]byte, error) {
	var value []byte = nil
	err := s.inst.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not found", table)
		}

		value = b.Get([]byte(key))
		return nil
	})

	return value, err
}

func (s *confManager) queryAll(table string, iterator func(k, v []byte) error) error {
	err := s.inst.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not found", table)
		}

		return b.ForEach(iterator)
	})

	return err
}

func (s *confManager) deleteOne(table string, key string) error {
	err := s.inst.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table %s not exist", table)
		}

		return b.Delete([]byte(key))
	})

	return err
}
