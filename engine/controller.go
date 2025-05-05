package engine

import (
	"os"
	"viction-rpc-crawler-go/config"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"

	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type Controller struct {
	Root   *config.RootConfig
	koanf  *koanf.Koanf
	Logger zerolog.Logger

	logFile *os.File
}

func NewController(useFS bool) *Controller {
	cfg, k, err := config.InitKoanf(useFS)
	logger, logFile, err2 := config.InitZerolog(cfg.ConfigDir, useFS)
	if err != nil {
		logger.Err(err).Msg("error initializing config")
	}
	if err2 != nil {
		logger.Err(err2).Msg("error initializing log file")
	}
	return &Controller{
		Root:   cfg,
		koanf:  k,
		Logger: logger,

		logFile: logFile,
	}
}

func (c *Controller) ConfigFromCli(cfgs map[string]interface{}) error {
	for k, v := range cfgs {
		c.koanf.Set(k, v)
	}
	err := c.koanf.Unmarshal("", &c.Root)
	return err
}

func (c *Controller) Close() {
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}
}

func (c *Controller) DbClient() (*db.DbClient, error) {
	return db.Connect(c.Root.Database.PostgreSQL, "")
}

func (c *Controller) RpcClient() (*rpc.EthClient, error) {
	return rpc.Connect(c.Root.Blockchain.RpcUrl)
}

func (c *Controller) CommandLogger(module, command string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Str("command", command).Logger()
}

func (c *Controller) ModuleLogger(module string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Logger()
}
