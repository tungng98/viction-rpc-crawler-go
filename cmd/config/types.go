package config

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	Crawler    *CrawlerConfig `koanf:"crawler"`
	MongoDB    *MongoDBConfig `koanf:"mongo"`
	Viction    *VictionConfig `koanf:"viction"`
}

type CrawlerConfig struct {
	StartBlock  uint64 `koanf:"startBlock"`
	EndBlock    uint64 `koanf:"endBlock"`
	BatchSize   int    `koanf:"batch"`
	WorkerCount int    `koanf:"worker"`
}

type MongoDBConfig struct {
	Host     string `koanf:"host"`
	Port     int32  `koanf:"port"`
	Username string `koanf:"usr"`
	Password string `koanf:"pwd"`
	Database string `koanf:"db"`
}

type VictionConfig struct {
	RpcUrl string `koanf:"rpc"`
}
