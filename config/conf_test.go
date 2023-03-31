package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func createDB() {
	_ = Init("test.db")
}

func destroyDB() {
	_ = Destroy()
}

func TestBool(t *testing.T) {
	createDB()

	assert.Equal(t, Bool("table", "ok"), false)

	SetBool("table", "ok", true)
	assert.Equal(t, Bool("table", "ok"), true)

	SetBool("table", "ok", false)
	assert.Equal(t, Bool("table", "ok"), false)

	Remove("table", "ok")
	assert.Equal(t, Bool("table", "ok"), false)

	SetString("table", "ok", "who knows?")
	t.Log("fetched string:", String("table", "ok"))
	assert.Equal(t, String("table", "ok"), "who knows?")

	//assert.Equal(t, String("table", "ok"), "who?")

	t.Log("TestBool done")

	destroyDB()
}
