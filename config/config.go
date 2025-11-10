package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Ethereum EthereumConfig `yaml:"ethereum"`
	TheGraph TheGraphConfig `yaml:"thegraph"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type EthereumConfig struct {
	RPCURL  string `yaml:"rpc_url"`
	Timeout string `yaml:"timeout"`
}

type TheGraphConfig struct {
	UniswapV2URL string  `yaml:"uniswap_v2_url"`
	MinTVL       float64 `yaml:"min_tvl"`
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
			Port: "1337",
		},
		Ethereum: EthereumConfig{
			RPCURL:  "",
			Timeout: "30s",
		},
		TheGraph: TheGraphConfig{
			UniswapV2URL: "",
			MinTVL:       10000.0, // $10,000 minimum TVL
		},
	}
}
