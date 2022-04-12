package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Entries   []*Entry
	Channels  map[string]string // channel name -> channel id
	Templates map[string]string
	Vars      map[string]interface{}
}

type Entry struct {
	Labels       []string
	Categories   []string
	Texts        []string
	Channels     []string
	Events       []*Event
	Template     string
	TemplateName string `yaml:"template_name"`
	Vars         map[string]interface{}
}

type Event struct {
	Type    string
	Actions []string
}

type configReader struct{}

type Reader interface {
	Read(p string, cfg *Config) error
}

func NewReader() Reader {
	return &configReader{}
}

func (reader *configReader) Read(p string, cfg *Config) error {
	f, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("open a configuration file %s: %w", p, err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return fmt.Errorf("parse a configuration file as YAML %s: %w", p, err)
	}
	return nil
}
