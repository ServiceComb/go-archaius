package archaius_test

import (
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-mesh/openlogging"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type EListener struct{}

func (e EListener) Event(event *event.Event) {
	openlogging.GetLogger().Infof("config value after change ", event.Key, " | ", event.Value)
}

var filename2 string

func TestInit(t *testing.T) {
	f1Bytes := []byte(`
age: 14
name: peter
`)
	f2Bytes := []byte(`
addr: somewhere
number: 1
exist: true
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f1.yaml")
	filename2 = filepath.Join(d, "f2.yaml")
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(filename1)
	f2, err := os.Create(filename2)
	assert.NoError(t, err)
	defer f2.Close()
	defer os.Remove(filename2)
	_, err = io.WriteString(f1, string(f1Bytes))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, string(f2Bytes))
	assert.NoError(t, err)
	os.Setenv("age", "15")
	err = archaius.Init(
		archaius.WithRequiredFiles([]string{filename1}),
		archaius.WithOptionalFiles([]string{filename2}),
		archaius.WithENVSource(),
		archaius.WithMemorySource())
	assert.NoError(t, err)
	assert.Equal(t, "15", archaius.Get("age"))
	t.Run("add mem config", func(t *testing.T) {
		archaius.Set("age", "16")
		assert.Equal(t, "16", archaius.Get("age"))
	})
	t.Run("delete mem config", func(t *testing.T) {
		archaius.Delete("age")
		assert.Equal(t, "15", archaius.Get("age"))
	})

}
func TestAddFile(t *testing.T) {

}
func TestConfig_Get(t *testing.T) {
	s := archaius.Get("number")
	assert.Equal(t, 1, s)

	e := archaius.GetBool("exist", false)
	assert.Equal(t, true, e)

	n := archaius.Get("name")
	assert.Equal(t, "peter", n)

	n3 := archaius.GetString("name", "")
	assert.Equal(t, "peter", n3)

	n2 := archaius.GetValue("name")
	name, err := n2.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "peter", name)

	b := archaius.Exist("name")
	assert.True(t, b)

	b = archaius.Exist("none")
	assert.False(t, b)

	m := archaius.GetConfigs()
	t.Log(m)
	assert.Equal(t, 1, m["number"])
}
func TestConfig_GetInt(t *testing.T) {
	s := archaius.GetInt("number", 0)
	assert.Equal(t, 1, s)
	s2 := archaius.GetInt64("number", 0)
	var a int64 = 1
	assert.Equal(t, a, s2)
}
func TestConfig_RegisterListener(t *testing.T) {
	eventHandler := EListener{}
	err := archaius.RegisterListener(eventHandler, "a*")
	assert.NoError(t, err)
	defer archaius.UnRegisterListener(eventHandler, "a*")

}

func TestUnmarshalConfig(t *testing.T) {
	b := []byte(`
key: peter
info:
  address: a
  number: 8
metadata_str:
  key01: "value01"
  key02: "value02"
metadata_int:
  key01: 1
  key02: 2
str_arr:
  - "list01"
  - "list02"
int_arr:
  - 1
  - 2
infos:
  - address: "addr01"
    number: 100
    users:
      - name: "yourname"
        age: 21
infos_ptr:
  - address: "addr02"
    number: 123
    users:
      - name: "yourname1"
        age: 22
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f3.yaml")
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	err = archaius.Init(archaius.WithMemorySource())
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(filename1)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	type User struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	type Info struct {
		Addr   string `yaml:"address"`
		Number int    `yaml:"number"`
		Us     []User `yaml:"users"`
	}
	type Person struct {
		Name     string            `yaml:"key"`
		MDS      map[string]string `yaml:"metadata_str"`
		MDI      map[string]int    `yaml:"metadata_int"`
		Info     *Info             `yaml:"info"`
		StrArr   []string          `yaml:"str_arr"`
		IntArr   []int             `yaml:"int_arr"`
		Infos    []Info            `yaml:"infos"`
		InfosPtr []*Info           `yaml:"infos_ptr"`
	}
	err = archaius.AddFile(filename1)
	assert.NoError(t, err)
	time.Sleep(time.Second * 3)
	p := &Person{}
	err = archaius.UnmarshalConfig(p)
	assert.NoError(t, err)
	assert.Equal(t, "peter", p.Name)
	assert.Equal(t, "value01", p.MDS["key01"])
	assert.Equal(t, 1, p.MDI["key01"])
	assert.Equal(t, "a", p.Info.Addr)
	assert.Equal(t, 8, p.Info.Number)
	assert.Equal(t, "list01", p.StrArr[0])
	assert.Equal(t, 1, p.IntArr[0])
	assert.Equal(t, "addr01", p.Infos[0].Addr)
	assert.Equal(t, "yourname", p.Infos[0].Us[0].Name)
	assert.Equal(t, "addr02", p.InfosPtr[0].Addr)
	assert.Equal(t, "yourname1", p.InfosPtr[0].Us[0].Name)
}
func TestInitConfigCenter(t *testing.T) {
	err := archaius.EnableRemoteSource("fake", nil)
	assert.Error(t, err)
}
func TestClean(t *testing.T) {
	err := archaius.Clean()
	assert.NoError(t, err)
	s := archaius.Get("age")
	assert.Equal(t, nil, s)
}
