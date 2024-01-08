package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidator(t *testing.T) {
	err := Load("test.json", "pareto", "root")
	assert.Nil(t, err)
	GetStore().Set("key1", 10)
	GetStore().Set("key2", 10.20)
	GetStore().Set("key3", -3)

	Clamp(GetStore(), "key1", GetStore().GetInt, 20, 30)
	Clamp(GetStore(), "key2", GetStore().GetFloat64, 5.0, 30)
	Clamp(GetStore(), "key3", GetStore().GetInt, -5, 30)

	assert.Equal(t, GetStore().GetInt("key1"), 20)
	assert.Equal(t, GetStore().GetFloat64("key2"), 10.2)
	assert.Equal(t, GetStore().GetInt("key3"), -3)
}

func TestDefault(t *testing.T) {
	err := Load("test.json", "pareto", "root")
	assert.Nil(t, err)

	assert.Equal(t, store.GetString("name"), "xxx")

	err = Load("agent.json", "service")
	assert.Nil(t, err)

	assert.Equal(t, store.GetString("name"), "agent")
}

func TestNewStore(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		store, err := NewStore("test.json", "pareto", "root")
		assert.Nil(t, err)

		assert.Equal(t, store.GetString("name"), "xxx")

		store.Set("service.registry", "nats://127.0.0.1:4222")

		err = store.WriteConfigAs("test.yml")
		assert.Nil(t, err)

	})

	t.Run("file-not-exist", func(t *testing.T) {
		store, err := NewStore("aaa.json")
		assert.NotNil(t, err)
		assert.Nil(t, store)

		t.Log(err)
	})
}
