package config

type Config struct {
	Server   ServerConfig   `env:"server"`
	Ethereum EthereumConfig `env:"ethereum"`
}
type ServerConfig struct {
	Port string `env:"port" envDefault:"1337"`
	Host string `env:"host" envDefault:"localhost"`
}

type EthereumConfig struct {
	RPCURL  string `env:"rpcURL" envDefault:"http://localhost:8545"`
	Timeout string `env:"timeout" envDefault:"30s"`
}
