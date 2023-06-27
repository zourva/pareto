package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadJsonConfig(t *testing.T) {
	file := "test.json"
	type testConf struct {
		Field1 string  `json:"field1,omitempty"`
		Field2 int     `json:"field2,omitempty"`
		Field3 float64 `json:"field3,omitempty"`
		Field4 int64   `json:"field4,omitempty"`
	}

	// gen config
	src := testConf{
		Field1: "field1:string",
		Field2: 2,
		Field3: 3.1,
		Field4: 4,
	}
	data, err := json.Marshal(&src)
	assert.Nil(t, err)
	err = os.WriteFile(file, data, 666)
	assert.Nil(t, err)

	var tc testConf
	err = LoadJsonConfig(file, &tc)
	assert.Nil(t, err)

	assert.Equal(t, src.Field1, tc.Field1)
	assert.Equal(t, src.Field2, tc.Field2)
	assert.Equal(t, src.Field3, tc.Field3)
	assert.Equal(t, src.Field4, tc.Field4)

	err = os.Remove(file)
	assert.Nil(t, err)
}
