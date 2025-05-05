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
	getBlocks.SetWorker(1)
	getBlocks.SetRouter(router)
	router.Register(getBlocks)

	if rpc == nil {
		logger.Warn("RPC services are not available.")
	} else {
		getBlock := NewGetBlock(logger, rpc)
		getBlock.SetWorker(cfg.Service.Worker.GetBlock)
		getBlock.SetRouter(router)
		router.Register(getBlock)
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
