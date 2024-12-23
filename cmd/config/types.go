package config

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Blockchain *BlockchainConfig `koanf:"blockchain"`
	Database   *DatabaseConfig   `koanf:"database"`
	ZeroLog    *ZeroLogConfig    `koanf:"zerolog"`
}

type DatabaseConfig struct {
	PostgreSQL string `koanf:"pgsql"`
}

type BlockchainConfig struct {
	RpcUrl string `koanf:"rpc"`
}

type ZeroLogConfig struct {
	Level        int8 `koanf:"level"`
	ConsoleLevel int8 `koanf:"consoleLevel"`
}
