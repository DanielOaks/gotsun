// Copyright (c) 2012-2014 Jeremy Latt
// Copyright (c) 2014-2015 Edmund Huber
// Copyright (c) 2016-2017 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package lib

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// TLSListenConfig defines configuration options for listening on TLS.
type TLSListenConfig struct {
	Cert string
	Key  string
}

// Config returns the TLS contiguration assicated with this TLSListenConfig.
func (conf *TLSListenConfig) Config() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(conf.Cert, conf.Key)
	if err != nil {
		return nil, errors.New("tls cert+key: invalid pair")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, err
}

// PromptsConfig holds the prompts that we tell the user on new diary entries.
type PromptsConfig struct {
	DefaultsEnabled []string `yaml:"defaults-enabled"`
	Lines           map[string][]string
}

// Config defines the overall configuration.
type Config struct {
	Listeners        []string
	TLSListenersInfo map[string]*TLSListenConfig `yaml:"tls-listeners"`
	StaticFiles      string                      `yaml:"static-files"`
	Templates        string                      `yaml:"templates"`
	Database         struct {
		Type string
		Path string
	}
	promptsPath string        `yaml:"prompts"`
	Prompts     PromptsConfig `yaml:"prompt-real"`
}

// TLSListeners returns a list of TLS listeners and their configs.
func (conf *Config) TLSListeners() map[string]*tls.Config {
	tlsListeners := make(map[string]*tls.Config)
	for s, tlsListenersConf := range conf.TLSListenersInfo {
		config, err := tlsListenersConf.Config()
		if err != nil {
			log.Fatal(err)
		}
		tlsListeners[s] = config
	}
	return tlsListeners
}

// LoadConfig loads the given YAML configuration file.
func LoadConfig(filename string) (config *Config, err error) {
	// load config file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// load prompts
	if config.promptsPath != "" {
		var promptsConfig *PromptsConfig

		data, err = ioutil.ReadFile(config.promptsPath)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, &promptsConfig)
		if err != nil {
			return nil, err
		}

		config.Prompts = *promptsConfig
	}

	return config, nil
}
