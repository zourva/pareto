package pareto

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/zourva/pareto/box"
	"github.com/zourva/pareto/box/env"
	"github.com/zourva/pareto/logger"
	"os"
	"testing"
)

// func TestSetupDefault(t *testing.T) {
//	Setup()
//	Teardown()
// }

func TestSetupWithOpts(t *testing.T) {
	options := []Option{
		WithLogger(
			logger.NewLogger(&logger.Options{
				Verbosity:   "vv",
				LogFileName: env.GetExecFilePath() + "/../log/22.log",
			}),
		),
		WithWorkingDir(
			env.NewWorkingDir(true,
				[]*env.DirInfo{
					{Name: "bin", Mode: 0755},
					{Name: "etc", Mode: 0755},
					{Name: "lib", Mode: 0755},
					{Name: "log", Mode: 0755},
					{Name: "data", Mode: 0755},
					{Name: "installer", Mode: 0755},
				}),
		),
	}

	SetupWithOpts(options...)
	Teardown()
}

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

func TestWithJsonConfParser(t *testing.T) {

	file := "test.json"
	type testConf struct {
		Field1 string  `json:"field1,omitempty"`
		Field2 int     `json:"field2,omitempty"`
		Field3 float64 `json:"field3,omitempty"`
		Field4 int64   `json:"field4,omitempty"`
		Field5 int     `json:"field5,omitempty"`
	}

	// gen config
	src := testConf{
		Field1: "field1:string",
		Field2: 2,
		Field3: 3.1,
		Field4: 4,
		Field5: 22,
	}
	data, err := json.Marshal(&src)
	assert.Nil(t, err)
	err = os.WriteFile(file, data, 666)
	assert.Nil(t, err)

	var tc testConf

	SetupWithOpts(WithJsonConfParser(file, &tc, func(obj any) error {
		o := obj.(*testConf)
		o.Field5 = 888
		return nil
	}))

	assert.Equal(t, src.Field1, tc.Field1)
	assert.Equal(t, src.Field2, tc.Field2)
	assert.Equal(t, src.Field3, tc.Field3)
	assert.Equal(t, src.Field4, tc.Field4)
	assert.Equal(t, 888, tc.Field5)

	err = os.Remove(file)
	assert.Nil(t, err)
}

func TestCreateLogger(t *testing.T) {
	SetupWithOpts(CreateLogger(func() *logger.Logger {
		return logger.NewLogger(&logger.Options{
			Verbosity:   "vv",
			LogFileName: env.GetExecFilePath() + "/../log/33.log",
		})
	}))

	assert.NotNil(t, bot.logger)
}

func TestCreateLogger2(t *testing.T) {

	file := "test.json"
	type testConf struct {
		Field1     string  `json:"field1,omitempty"`
		Field2     int     `json:"field2,omitempty"`
		Field3     float64 `json:"field3,omitempty"`
		Field4     int64   `json:"field4,omitempty"`
		Field5     int     `json:"field5,omitempty"`
		LoggerFile string  `json:"loggerFile,omitempty"`
	}

	// gen config
	src := testConf{
		Field1:     "field1:string",
		Field2:     2,
		Field3:     3.1,
		Field4:     4,
		Field5:     22,
		LoggerFile: "test.log",
	}
	data, err := json.Marshal(&src)
	assert.Nil(t, err)
	err = os.WriteFile(file, data, 666)
	assert.Nil(t, err)

	var tc testConf
	SetupWithOpts(
		WithJsonConfParser(file, &tc, nil),
		CreateLogger(func() *logger.Logger {
			return logger.NewLogger(&logger.Options{
				Verbosity:   "vv",
				LogFileName: env.GetExecFilePath() + tc.LoggerFile,
			})
		}))
	log.Info("this is a test log line...")

	ok, err := box.PathExists(env.GetExecFilePath() + tc.LoggerFile)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	err = os.Remove(file)
	assert.Nil(t, err)
}
