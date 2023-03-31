package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	err := Init("conf.db")
	assert.Nil(t, err)

	Destroy()

	// multiple init
	err = Init("conf.db")
	assert.Nil(t, err)

	err = Init("conf.db")
	assert.Nil(t, err)

	err = Init("conf.db")
	assert.Nil(t, err)

	Destroy()
}
