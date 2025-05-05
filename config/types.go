package config

const (
	BlockchainRpcUrlKey = "blockchain.rpc"
)

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Blockchain *BlockchainConfig `koanf:"blockchain"`
	Database   *DatabaseConfig   `koanf:"database"`
	ZeroLog    *ZeroLogConfig    `koanf:"zerolog"`
	Service    *ServiceConfig    `koanf:"service"`
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

type ServiceConfig struct {
	Schedule *ServiceScheduleConfig `koanf:"schedule"`
	Worker   *JobWorkerConfig       `koanf:"worker"`
}

type ServiceScheduleConfig struct {
	IndexBlockInterval int64 `koanf:"indexBlockInterval"`
	IndexBlockBatch    int   `koanf:"indexBlockBatch"`
}

type JobWorkerConfig struct {
	GetBlock uint64 `koanf:"getBlock"`
}
