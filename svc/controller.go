package svc

import (
	"viction-rpc-crawler-go/config"
	"viction-rpc-crawler-go/db"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type Controller struct {
	cfg    *config.RootConfig
	db     *db.DbClient
	rpc    *rpc.EthClient
	svc    *multiplex.ServiceController
	logger diag.Logger
}

func NewController(cfg *config.RootConfig, db *db.DbClient, rpc *rpc.EthClient, logger diag.Logger) *Controller {
	router := multiplex.NewServiceController(logger)

	getBlocks := NewGetBlocks(logger)
	getBlocks.SetRouter(router)
	getBlocks.SetWorker(1)
	router.Register(getBlocks)

	traceBlocks := NewTraceBlocks(logger)
	traceBlocks.SetRouter(router)
	traceBlocks.SetWorker(1)
	router.Register(traceBlocks)

	downloadBlock := NewDownloadBlock(logger)
	downloadBlock.SetRouter(router)
	downloadBlock.SetWorker(4)
	router.Register(downloadBlock)

	readFileSystem := NewReadFileSystem(logger)
	readFileSystem.SetRouter(router)
	readFileSystem.SetWorker(1)
	router.Register(readFileSystem)

	writeFileSystem := NewWriteFileSystem(logger)
	writeFileSystem.SetRouter(router)
	writeFileSystem.SetWorker(1)
	router.Register(writeFileSystem)

	if db == nil {
		logger.Warn("DB services are not available.")
	} else {
		readDatabase := NewReadDatabase(logger, db)
		readDatabase.SetRouter(router)
		readDatabase.SetWorker(1)
		router.Register(readDatabase)

		writeDatabase := NewWriteDatabase(logger, db)
		writeDatabase.SetRouter(router)
		writeDatabase.SetWorker(1)
		router.Register(writeDatabase)
	}

	if rpc == nil {
		logger.Warn("RPC services are not available.")
	} else {
		getBlock := NewGetBlock(logger, rpc)
		getBlock.SetRouter(router)
		getBlock.SetWorker(cfg.Service.Worker.GetBlock)
		router.Register(getBlock)

		traceBlock := NewTraceBlock(logger, rpc)
		traceBlock.SetRouter(router)
		traceBlock.SetWorker(cfg.Service.Worker.GetBlock)
		router.Register(traceBlock)
	}

	return &Controller{
		cfg:    cfg,
		db:     db,
		rpc:    rpc,
		svc:    router,
		logger: logger,
	}
}

func (c *Controller) Run() {
	c.svc.Run(true)
}

func (c *Controller) Dispatch(serviceID string, command string, params multiplex.ExecParams) {
	c.svc.Dispatch(serviceID, command, params)
}

func (c *Controller) DispatchOnce(serviceID string, command string, params multiplex.ExecParams) {
	params.ExpectReturn()
	c.svc.Dispatch(serviceID, command, params)
	params.Wait()
	c.svc.Exec("exit", multiplex.ExecParams{})
}
