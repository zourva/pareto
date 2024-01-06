package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
