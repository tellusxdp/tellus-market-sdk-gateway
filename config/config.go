package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type HTTP struct {
	ListenAddress string `yaml:"listen_address"`
	TLS           *struct {
		Autocert    bool   `yaml:"autocert"`
		Certificate string `yaml:"certificate,omitempty"`
		Key         string `yaml:"key,omitempty"`
	} `yaml:"tls,omitempty"`
}

type Upstream struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

type Config struct {
	HTTP          HTTP     `yaml:"http"`
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
