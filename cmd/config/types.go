package config

type RootConfig struct {
	ConfigDir  string
	ConfigFile string
	IsPortable bool
	MongoDB    *MongoDBConfig `koanf:"mongo"`
	Viction    *VictionConfig `koanf:"viction"`
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
