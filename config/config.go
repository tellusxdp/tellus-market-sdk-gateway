package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Upstream struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

type Config struct {
	ListenAddress string   `yaml:"listen_address"`
	PrivateKeyURL string   `yaml:"private_key_url"`
	CounterURL    string   `yaml:"counter_url"`
	APIKey        string   `yaml:"api_key"`
	Upstream      Upstream `yaml:"upstream"`
	ProviderName  string   `yaml:"provider_name"`
	ToolLabel     string   `yaml:"tool_label"`
	ToolID        string   `yaml:"tool_id"`
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

	if c.Upstream.Headers == nil {
		c.Upstream.Headers = map[string]string{}
	}

	return c, err
}
