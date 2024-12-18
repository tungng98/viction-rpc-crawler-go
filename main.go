package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"viction-rpc-crawler-go/cmd"
	"viction-rpc-crawler-go/cmd/config"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/diag"
	"viction-rpc-crawler-go/svc"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"
)

var invokeArgs cmd.Args

func main() {
	cfg, cfgErr := config.InitKoanf()
	logFile := diag.InitZerolog(cfg.ConfigDir)
	if logFile != nil {
		defer logFile.Close()
	}

	invokeArgs = cmd.Args{}
	arg.MustParse(&invokeArgs)
	pwd, _ := os.Getwd()
	pwd, _ = filepath.Abs(pwd)
	pwd = strings.ReplaceAll(pwd, "\\", "/")
	exec, _ := os.Executable()
	exec, _ = filepath.Abs(exec)
	exec = strings.ReplaceAll(exec, "\\", "/")
	log.Info().Msg("VICTION CRAWLER")
	log.Info().Msgf("Working directory %s", pwd)
	log.Info().Msgf("Config directory %s", cfg.ConfigDir)
	log.Info().Msgf("Config file %s", cfg.ConfigFile)
	log.Info().Msgf("Executable file %s", exec)
	log.Info().Msgf("Portable mode %t", cfg.IsPortable)
	if cfgErr != nil {
		panic(cfgErr)
	}
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s", cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Host, cfg.MongoDB.Port, cfg.MongoDB.Database)
	if cfg.MongoDB.Username == "" || cfg.MongoDB.Password == "" {
		connStr = fmt.Sprintf("mongodb://%s:%d", cfg.MongoDB.Host, cfg.MongoDB.Port)
	}

	if invokeArgs.IndexBlockTx != nil {
		indexCfg := invokeArgs.IndexBlockTx
		svc := &svc.IndexBlockTxService{
			DbConnStr:       connStr,
			RpcUrl:          cfg.Viction.RpcUrl,
			Logger:          &log.Logger,
			BatchSize:       int(indexCfg.BatchSize),
			WorkerCount:     int(indexCfg.WorkerCount),
			StartBlock:      int64(indexCfg.StartBlock),
			EndBlock:        int64(indexCfg.EndBlock),
			UseHighestBlock: !indexCfg.Forced,
		}
		svc.Exec()
	}
	if invokeArgs.ScanBlockForError != nil {
		traceCfg := invokeArgs.ScanBlockForError
		svc := &svc.TraceBlockService{
			DbConnStr:          connStr,
			DbName:             cfg.MongoDB.Database,
			RpcUrl:             cfg.Viction.RpcUrl,
			Logger:             &log.Logger,
			WorkerCount:        int(traceCfg.WorkerCount),
			BatchSize:          int(traceCfg.BatchSize),
			StartBlock:         int64(traceCfg.StartBlock),
			EndBlock:           int64(traceCfg.EndBlock),
			UseCheckpointBlock: !traceCfg.NoCheckpoint,
			SaveDebugData:      !traceCfg.NoSaveTrace,
		}
		svc.Exec()
	}
	if invokeArgs.ManageDatabase != nil {
		subArgs := invokeArgs.ManageDatabase
		if subArgs.Migrate != nil {
			c, err := db.Connect(subArgs.Migrate.PostgreSQL, "")
			if err != nil {
				log.Error().Err(err).Msg("Cannot connect to database")
				return
			}
			c.Migrate()
		}
	}
}
