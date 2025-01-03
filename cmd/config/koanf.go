package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

var cfg *RootConfig

func BuildConfig(f string) (*RootConfig, error) {
	k := defaultConfig()
	if isExist(f) {
		k, _ = configFromYaml(k, f)
	}
	k, _ = configFromEnv(k)

	var config RootConfig
	err := k.Unmarshal("", &config)
	return &config, err
}

func InitKoanf() (*RootConfig, error) {
	if cfg != nil {
		return cfg, nil
	}
	configFile := "rpc-crawler.yml"
	exec, _ := os.Executable()
	exec, _ = filepath.Abs(exec)
	execName := filepath.Base(exec)
	cfgName := strings.TrimSuffix(execName, filepath.Ext(execName)) + ".yml"
	if strings.HasPrefix(execName, "__debug_bin") {
		cfgName = "viction-rpc-crawler-go.yml"
	}
	isPortable := false
	if isExist(path.Join(path.Dir(exec), cfgName)) {
		configFile = path.Join(path.Dir(exec), cfgName)
		isPortable = true
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = path.Join(home, ".config", "vicsvc", cfgName)
	} else if runtime.GOOS == "windows" {
		appData := strings.ReplaceAll(os.Getenv("APPDATA"), "\\", "/")
		configFile = path.Join(appData, "VicSvc", cfgName)
	}
	var err error
	cfg, err = BuildConfig(configFile)
	if err != nil {
		return cfg, err
	}

	cfg.ConfigDir = path.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, nil
}

func defaultConfig() *koanf.Koanf {
	var k = koanf.New(".")

	k.Load(
		structs.Provider(RootConfig{
			Blockchain: &BlockchainConfig{
				RpcUrl: "http://localhost:8545",
			},
			Database: &DatabaseConfig{},
			ZeroLog: &ZeroLogConfig{
				Level:        int8(zerolog.DebugLevel),
				ConsoleLevel: int8(zerolog.DebugLevel),
			},
			Service: &ServiceConfig{
				Schedule: &ServiceScheduleConfig{},
				Worker: &JobWorkerConfig{
					BlockFetcher: 16,
				},
			},
		}, "koanf"),
		nil,
	)

	return k
}

func configFromEnv(k *koanf.Koanf) (*koanf.Koanf, error) {
	err := k.Load(env.Provider("VICSVC_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(s, "VICSVC_")), "_", ".", -1)
	}), nil)
	if err != nil {
		return k, err
	}
	return k, nil
}

func configFromYaml(k *koanf.Koanf, f string) (*koanf.Koanf, error) {
	err := k.Load(file.Provider(f), yaml.Parser())
	if err != nil {
		return k, err
	}
	return k, nil
}

func isExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return !os.IsNotExist(err)
}
