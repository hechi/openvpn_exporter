package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	sc := &SafeConfig{
		C: &List{},
	}

	err := sc.Load("testdata/good.yaml")
	if err != nil {
		t.Errorf("Error loading config %v: %v", "good.yml", err)
	}

	c := sc.C

	assert.NoError(t, err)
	assert.Len(t, c.Config, 3)
	assert.Equal(t, c.Config[0].Name, "server2")
	assert.Equal(t, c.Config[0].LogFile, "examples/server2.status")
	assert.Equal(t, c.Config[1].Name, "server3")
	assert.Equal(t, c.Config[2].Name, "client")
}

func TestLoadBadConfigs(t *testing.T) {
	sc := &SafeConfig{
		C: &List{},
	}
	tests := []struct {
		ConfigFile    string
		ExpectedError string
	}{
		{
			ConfigFile:    "testdata/bad_config1.yaml",
			ExpectedError: "error parsing config file: yaml: unmarshal errors:\n  line 2: cannot unmarshal !!map into []config.Config",
		},
		{
			ConfigFile:    "testdata/bad_config2.yaml",
			ExpectedError: "error parsing config file: config:logfile is required",
		},
		{
			ConfigFile:    "testdata/bad_config3.yaml",
			ExpectedError: "error parsing config file: config:name is required",
		},
		{
			ConfigFile:    "testdata/bad_config4.yaml",
			ExpectedError: "error parsing config file: yaml: unmarshal errors:\n  line 4: field blah not found in type config.plain",
		},
	}
	for i, test := range tests {
		err := sc.Load(test.ConfigFile)
		if err == nil {
			t.Errorf("In case %v:\nExpected:\n%v\nGot:\nnil", i, test.ExpectedError)
			continue
		}
		if err.Error() != test.ExpectedError {
			t.Errorf("In case %v:\nExpected:\n%v\nGot:\n%v", i, test.ExpectedError, err.Error())
		}
	}
}
