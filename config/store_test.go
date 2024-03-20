package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidator(t *testing.T) {
	err := GetStore().Load("test.json", Json, "pareto", "root")
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
	err := GetStore().Load("test.json", Json, "pareto", "root")
	assert.Nil(t, err)

	t.Log(GetStore().All())

	assert.Equal(t, GetString("name"), "xxx")

	err = GetStore().Load("agent.json", Json, "service")
	assert.Nil(t, err)

	t.Log(GetStore().All())

	assert.Equal(t, GetString("name"), "agent")
	assert.Equal(t, GetString("keep"), "KeepMe")
}

//func TestNewStore(t *testing.T) {
//	t.Run("normal", func(t *testing.T) {
//		store, err := NewStore("test.json", "pareto", "root")
//		assert.Nil(t, err)
//
//		assert.Equal(t, store.GetString("name"), "xxx")
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
