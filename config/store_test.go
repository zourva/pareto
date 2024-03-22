package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidator(t *testing.T) {
	err := GetStore().Load("test.json", Text, Json, "pareto", "root")
	assert.Nil(t, err)
	GetStore().Set("key1", 10)
	GetStore().Set("key2", 10.20)
	GetStore().Set("key3", -3)

	Clamp(GetStore(), "key1", GetStore().Int, 20, 30)
	Clamp(GetStore(), "key2", GetStore().Float64, 5.0, 30)
	Clamp(GetStore(), "key3", GetStore().Int, -5, 30)

	assert.Equal(t, GetStore().Int("key1"), 20)
	assert.Equal(t, GetStore().Float64("key2"), 10.2)
	assert.Equal(t, GetStore().Int("key3"), -3)
}

func TestDefault(t *testing.T) {
	err := GetStore().Load("test.json", Text, Json, "pareto", "root")
	assert.Nil(t, err)

	t.Log(GetStore().All())

	assert.Equal(t, String("name"), "xxx")

	err = GetStore().Load("agent.json", Text, Json, "service")
	assert.Nil(t, err)

	t.Log(GetStore().All())

	assert.Equal(t, String("name"), "agent")
	assert.Equal(t, String("keep"), "KeepMe")
}

func TestBoltdb(t *testing.T) {
	s := newStorage("test.db")
	require.Nil(t, s.open())

	err := s.upsert("global", "config", []byte(`{
  "lwm2m":{
	"oma":{},
	"record":{
    	"obj00":[{"bn":"/0/0/","n":"0","vs":"xxx:5684"},{"n":"1","vb":true}],
    	"obj01":[{"bn":"/0/1/","n":"0","vs":"xxx:5684"},{"n":"1","vb":false}]
	}
  },
  "agent":{
    "monitor":{
      "healthCheck":"http://localhost:8222/healthz",
      "networkCheck": {
		"endpoint": "https://ifconfig.me/ip"
      }
    },
    "courier":{
      "name":"some-name",
      "interval": 500,
      "timeout":  100
    }
  }
}
`))
	require.Nil(t, err)
	require.Nil(t, s.close())

	// test load
	dbStore := New()
	require.Nil(t, dbStore.Load("test.db", Boltdb, Json))
	for k, v := range dbStore.All() {
		t.Log(k, ":", v)
	}

	// test sub-store
	sub := dbStore.Cut("agent")
	sub.Set("courier.timeout", 10)
	sub.Set("newlyadd", "added")
	for k, v := range sub.All() {
		t.Log("sub:", k, ":", v)
	}

	// test merge
	dbStore.MergeAt(sub, "agent")
	assert.Equal(t, 10, dbStore.Int("agent.courier.timeout"))
	assert.Equal(t, "added", dbStore.String("agent.newlyadd"))
	t.Log(dbStore.Get("lwm2m.record.obj00"))
	v := dbStore.Get("lwm2m.record.obj00")
	v0 := v.([]any)[0]
	t.Log(v0.(map[string]any)["n"])
	//assert.Equal(t, "0", dbStore.String("lwm2m.obj00.0.n"))
	//assert.Equal(t, "xxx:5684", dbStore.String("lwm2m.obj00.0.vs"))

	//os.Remove("test.db")

}

//func TestNewStore(t *testing.T) {
//	t.Run("normal", func(t *testing.T) {
//		store, err := NewStore("test.json", "pareto", "root")
//		assert.Nil(t, err)
//
//		assert.Equal(t, store.String("name"), "xxx")
//
//		store.Set("service.registry", "nats://127.0.0.1:4222")
//
//		err = store.WriteConfigAs("test.yml")
//		assert.Nil(t, err)
//
//	})
//
//	t.Run("file-not-exist", func(t *testing.T) {
//		store, err := NewStore("aaa.json")
//		assert.NotNil(t, err)
//		assert.Nil(t, store)
//
//		t.Log(err)
//	})
//}
