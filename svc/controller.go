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
	svcs   map[string]multiplex.Service
	logger diag.Logger
}

func NewController(cfg *config.RootConfig, db *db.DbClient, rpc *rpc.EthClient, logger diag.Logger) *Controller {
	router := multiplex.NewServiceController(logger)

	svcs := map[string]multiplex.Service{}
	getBlocks := NewGetBlocks(logger)
	getBlocks.SetWorker(1)
	getBlocks.SetRouter(router)
	router.Register(getBlocks)
	svcs[getBlocks.ServiceID()] = getBlocks

	if rpc == nil {
		logger.Warn("RPC services are not available.")
	} else {
		getBlock := NewGetBlock(logger, rpc)
		getBlock.SetWorker(cfg.Service.Worker.GetBlock)
		getBlock.SetRouter(router)
		router.Register(getBlock)
		svcs[getBlock.ServiceID()] = getBlock
	}

	return &Controller{
		cfg:    cfg,
		db:     db,
		rpc:    rpc,
		svc:    router,
		svcs:   svcs,
		logger: logger,
	}
}

func (c *Controller) Run() {
	c.svc.Run(true)
}

func (c *Controller) Dispatch(serviceID string, command string, params multiplex.ExecParams) {
	c.svcs[serviceID].Exec(command, params)
}

func (c *Controller) DispatchOnce(serviceID string, command string, params multiplex.ExecParams) {
	params.ExpectReturn()
	c.svcs[serviceID].Exec(command, params)
	params.Wait()
	c.svc.Exec("exit", multiplex.ExecParams{})
}
