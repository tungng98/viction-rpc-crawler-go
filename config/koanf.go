package config

import (
	"os"
	"path"
	"runtime"
	"strings"
	"viction-rpc-crawler-go/filesystem"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/strfmt"
)

var cfg *RootConfig
var k *koanf.Koanf

func BuildConfig(useFS bool, f string) (*RootConfig, *koanf.Koanf, error) {
	k := defaultConfig()
	if useFS && filesystem.IsFileExist(f) {
		k, _ = configFromYaml(k, f)
	}
	k, _ = configFromEnv(k)

	var config RootConfig
	err := k.Unmarshal("", &config)
	return &config, k, err
}

func InitKoanf(useFS bool) (*RootConfig, *koanf.Koanf, error) {
	if cfg != nil {
		return cfg, k, nil
	}
	configFile := "viction-crawler.yml"
	exec, _ := os.Executable()
	execPath := strfmt.NewPathFromStr(exec)
	cfgName := execPath.Name.Name + ".yml"
	if strings.HasPrefix(execPath.Name.Name, "__debug_bin") {
		cfgName = "viction-crawler-debug.yml"
	}
	isPortable := IsPortable()
	if isPortable {
		configFile = path.Join(execPath.ParentPath(), cfgName)
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		home := os.Getenv("HOME")
		configFile = path.Join(home, ".config", "vicsvc", cfgName)
	} else if runtime.GOOS == "windows" {
		appData := strings.ReplaceAll(os.Getenv("APPDATA"), "\\", "/")
		configFile = path.Join(appData, "VicSvc", cfgName)
	}
	var err error
	cfg, k, err = BuildConfig(useFS, configFile)
	if err != nil {
		return cfg, k, err
	}

	cfg.ConfigDir = path.Dir(configFile)
	cfg.ConfigFile = configFile
	cfg.IsPortable = isPortable
	return cfg, k, nil
}

func ExecPath() *strfmt.Path {
	exec, _ := os.Executable()
	execPath := strfmt.NewPathFromStr(exec)
	return execPath
}

func IsPortable() bool {
	exec, _ := os.Executable()
	configPath := strfmt.NewPathFromStr(exec)
	configPath.Name.Extension = ".yml"
	return filesystem.IsFileExist(configPath.FullPath())
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
