package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Autocert struct {
	Enabled  bool   `yaml:"enabled"`
	CacheDir string `yaml:"cache_dir"`
}

type HTTP struct {
	ListenAddress string `yaml:"listen_address"`
	TLS           *struct {
		Autocert    *Autocert `yaml:"autocert,omitempty"`
		Certificate string    `yaml:"certificate,omitempty"`
		Key         string    `yaml:"key,omitempty"`
	} `yaml:"tls,omitempty"`
}

type Upstream struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

type Config struct {
	HTTP             HTTP     `yaml:"http"`
	PrivateKeyURL    string   `yaml:"private_key_url"`
	CounterURL       string   `yaml:"counter_url"`
	APIKey           string   `yaml:"api_key"`
	Upstream         Upstream `yaml:"upstream"`
	ProviderId	     string   `yaml:"provider_id"`
	ToolLabel        string   `yaml:"tool_label"`
	ToolID           string   `yaml:"tool_id"`
	AllowedAuthTypes []string `yaml:"allowd_auth_types"`
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

	if c.HTTP.TLS != nil && c.HTTP.TLS.Autocert == nil {
		c.HTTP.TLS.Autocert = &Autocert{Enabled: false}
	}

	if c.AllowedAuthTypes == nil {
		c.AllowedAuthTypes = []string{}
	}

	return c, err
}
