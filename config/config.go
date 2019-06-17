package config

import (
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type Config struct {
	Name    string `yaml:"name"`
	LogFile string `yaml:"logfile"`
}

type List struct {
	Config []Config `yaml:"configs"`
}

type SafeConfig struct {
	sync.RWMutex
	C *List
}

func (sc *SafeConfig) Load(confFile string) (err error) {
	var c = &List{}

	yamlReader, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	log.Infof("config: %v", c)

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *List) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain List
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.Name == "" {
		return errors.New("config:name is required")
	}
	if c.LogFile == "" {
		return errors.New("config:logfile is required")
	}
	return nil
}
