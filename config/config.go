package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config struct {
	ListenAddress string `yaml:"listen_address"`
	PrivateKeyURL string `yaml:"private_key_url"`
	Upstream      string `yaml:"upstream"`
	ProviderName  string `yaml:"provider_name"`
	ToolLabel     string `yaml:"tool_label"`
	ToolID        string `yaml:"tool_id"`
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
