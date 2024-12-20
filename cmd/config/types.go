package config

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Blockchain *BlockchainConfig `koanf:"blockchain"`
	Database   *DatabaseConfig   `koanf:"database"`
}

type DatabaseConfig struct {
	PostgreSQL string `koanf:"pgsql"`
}

type BlockchainConfig struct {
	RpcUrl string `koanf:"rpc"`
}
