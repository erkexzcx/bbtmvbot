package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Telegram struct {
		ApiKey string `yaml:"api_key"`
	} `yaml:"telegram"`
}

func New(path string) (*Config, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(contents, &c)
	return &c, err
}
