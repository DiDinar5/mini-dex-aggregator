package config

import (
	"fmt"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Ethereum EthereumConfig `yaml:"ethereum"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type EthereumConfig struct {
	RPCURL  string `yaml:"rpc_url"` //the graph
	Timeout string `yaml:"timeout"`
}

func Load() *Config {
	config, err := loadFromYAML("config.yaml")
	if err != nil {
		return defaultConfig()
	}

	return config
}

func loadFromYAML(filename string) (*Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s not found", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &config, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: ":1337",
		},
		Ethereum: EthereumConfig{
			RPCURL:  "http://localhost:8545",
			Timeout: "30s",
		},
	}
}
