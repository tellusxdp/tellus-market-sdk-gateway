package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	ListenAddress string `yaml:"listen_address"`
	PrivateKeyURL string `yaml:"private_key_url"`
	Upstream      string `yaml:"upstream"`
	ProductID     string `yaml:"product"`
}

func FromFilepath(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, err
}
