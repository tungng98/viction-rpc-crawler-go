package config

import "github.com/rs/zerolog"

const (
	BlockchainRpcUrlKey   = "blockchain.rpc"
	FileSystemRootPathKey = "filesystem.rootPath"
)

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Blockchain *BlockchainConfig `koanf:"blockchain"`
	Database   *DatabaseConfig   `koanf:"database"`
	FileSystem *FileSystemConfig `koanf:"filesystem"`
	ZeroLog    *ZeroLogConfig    `koanf:"zerolog"`
	Service    *ServiceConfig    `koanf:"service"`
}

type BlockchainConfig struct {
	RpcUrl string `koanf:"rpc"`
}

type DatabaseConfig struct {
	PostgreSQL string `koanf:"pgsql"`
}

type FileSystemConfig struct {
	RootPath string `koanf:"rootPath"`
}

type ZeroLogConfig struct {
	Level        int8 `koanf:"level"`
	ConsoleLevel int8 `koanf:"consoleLevel"`
}

type ServiceConfig struct {
	Schedule *SchedulerConfig `koanf:"schedule"`
	Worker   *JobWorkerConfig `koanf:"worker"`
}

type SchedulerConfig struct {
	IndexBlockInterval int64 `koanf:"indexBlockInterval"`
	IndexBlockBatch    int   `koanf:"indexBlockBatch"`
}

type JobWorkerConfig struct {
	GetBlock uint64 `koanf:"getBlock"`
}

func DefaultRootConfig() *RootConfig {
	return &RootConfig{
		Blockchain: &BlockchainConfig{
			RpcUrl: "http://localhost:8545",
		},
		Database:   &DatabaseConfig{},
		FileSystem: &FileSystemConfig{},
		ZeroLog: &ZeroLogConfig{
			Level:        int8(zerolog.DebugLevel),
			ConsoleLevel: int8(zerolog.DebugLevel),
		},
		Service: &ServiceConfig{
			Schedule: &SchedulerConfig{
				IndexBlockInterval: 30,
				IndexBlockBatch:    900,
			},
			Worker: &JobWorkerConfig{
				GetBlock: 8,
			},
		},
	}
}
