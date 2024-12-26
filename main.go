package main

import (
	"os"
	"path/filepath"
	"strings"
	"viction-rpc-crawler-go/cache"
	"viction-rpc-crawler-go/cmd"
	"viction-rpc-crawler-go/cmd/config"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/diag"
	"viction-rpc-crawler-go/rpc"
	"viction-rpc-crawler-go/svc"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var invokeArgs cmd.Args

func main() {
	cfg, cfgErr := config.InitKoanf()
	logFile := diag.InitZerolog(cfg.ConfigDir, cfg.ZeroLog.Level, cfg.ZeroLog.ConsoleLevel)
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
	log.Info().Msgf("Working directory: %s", pwd)
	log.Info().Msgf("Config directory: %s", cfg.ConfigDir)
	log.Info().Msgf("Config file: %s", cfg.ConfigFile)
	log.Info().Msgf("Executable file: %s", exec)
	log.Info().Msgf("Portable mode: %t", cfg.IsPortable)
	if cfgErr != nil {
		panic(cfgErr)
	}

	if invokeArgs.IndexBlockTx != nil {
		indexCfg := invokeArgs.IndexBlockTx
		if indexCfg.RpcUrl != "" {
			cfg.Blockchain.RpcUrl = indexCfg.RpcUrl
		}
		if indexCfg.PostgreSQL != "" {
			cfg.Database.PostgreSQL = indexCfg.PostgreSQL
		}
		svc := &svc.IndexBlockTxService{
			DbConnStr:       cfg.Database.PostgreSQL,
			RpcUrl:          cfg.Blockchain.RpcUrl,
			Logger:          &log.Logger,
			BatchSize:       int(indexCfg.BatchSize),
			WorkerCount:     int(indexCfg.WorkerCount),
			StartBlock:      int64(indexCfg.StartBlock),
			EndBlock:        int64(indexCfg.EndBlock),
			UseHighestBlock: !indexCfg.Forced,
			IncludeTxs:      indexCfg.IncludeTxs,
		}
		svc.Exec()
	}
	if invokeArgs.ManageDatabase != nil {
		subArgs := invokeArgs.ManageDatabase
		if subArgs.Migrate.PostgreSQL != "" {
			cfg.Database.PostgreSQL = subArgs.Migrate.PostgreSQL
		}
		if subArgs.Migrate != nil {
			c, err := db.Connect(cfg.Database.PostgreSQL, "")
			if err != nil {
				log.Error().Err(err).Msg("Cannot connect to database.")
				return
			}
			err = c.Migrate()
			if err != nil {
				log.Error().Err(err).Msg("Error while migrating database.")
				return
			}
			log.Info().Msg("Migration successful!")
		}
	}
	getModuleLogger := func(name string) *zerolog.Logger {
		moduleLogger := log.Logger.With().Str("Module", name).Logger()
		return &moduleLogger
	}
	if invokeArgs.Service != nil {
		controller := svc.NewServiceController(getModuleLogger("Controller"))
		rpcClient, err := rpc.Connect(cfg.Blockchain.RpcUrl)
		if err != nil {
			log.Error().Err(err).Str("Url", cfg.Blockchain.RpcUrl).Msg("Cannot connect to RPC")
			return
		}
		dbClient, err := db.Connect(cfg.Database.PostgreSQL, "")
		if err != nil {
			log.Error().Err(err).Str("Url", cfg.Database.PostgreSQL).Msg("Cannot connect to database")
			return
		}
		memCache := cache.NewMemoryCache()
		scheduler := svc.NewScheduleSvc(250, controller, getModuleLogger("Scheduler"))
		controller.RegisterService(scheduler)
		blockFetcher := svc.NewBlockFetcherSvc(controller, rpcClient, memCache, getModuleLogger("BlockFetcher"))
		controller.RegisterService(blockFetcher)
		blockIndexer := svc.NewBlockIndexerSvc(controller, rpcClient, memCache, dbClient, getModuleLogger("BlockIndexer"))
		controller.RegisterService(blockIndexer)

		controller.Exec("set_worker", map[string]interface{}{"service_id": scheduler.ServiceID(), "worker_count": uint16(1)})
		controller.Exec("set_worker", map[string]interface{}{"service_id": blockFetcher.ServiceID(), "worker_count": uint16(8)})
		controller.Exec("set_worker", map[string]interface{}{"service_id": blockIndexer.ServiceID(), "worker_count": uint16(1)})
		controller.Run(true)
	}
}
