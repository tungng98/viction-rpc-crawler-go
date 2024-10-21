package main

import (
	"fmt"
	"os"
	"path/filepath"
	"viction-rpc-crawler-go/cmd/config"
	"viction-rpc-crawler-go/diag"
	"viction-rpc-crawler-go/svc"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, cfgErr := config.InitKoanf()
	logFile := diag.InitZerolog(cfg.ConfigDir)
	if logFile != nil {
		defer logFile.Close()
	}

	pwd, _ := os.Getwd()
	pwd, _ = filepath.Abs(pwd)
	exec, _ := os.Executable()
	exec, _ = filepath.Abs(exec)
	log.Info().Msg("VICTION CRAWLER")
	log.Info().Msgf("Working directory %s", pwd)
	log.Info().Msgf("Config directory %s", cfg.ConfigDir)
	log.Info().Msgf("Executable file %s", exec)
	log.Info().Msgf("Portable mode %t", cfg.IsPortable)
	if cfgErr != nil {
		panic(cfgErr)
	}
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s", cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Host, cfg.MongoDB.Port, cfg.MongoDB.Database)
	if cfg.MongoDB.Username == "" || cfg.MongoDB.Password == "" {
		connStr = fmt.Sprintf("mongodb://%s:%d", cfg.MongoDB.Host, cfg.MongoDB.Port)
	}
	svc := &svc.IndexBlockTxService{
		DbConnStr:   connStr,
		DbName:      cfg.MongoDB.Database,
		RpcUrl:      cfg.Viction.RpcUrl,
		Logger:      &log.Logger,
		BatchSize:   cfg.Crawler.BatchSize,
		WorkerCount: cfg.Crawler.WorkerCount,
		StartBlock:  int64(cfg.Crawler.StartBlock),
		EndBlock:    int64(cfg.Crawler.EndBlock),
	}
	svc.Exec()
}
