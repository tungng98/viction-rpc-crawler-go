package svc

import (
	"viction-rpc-crawler-go/db"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type ReadDatabase struct {
	multiplex.ServiceCore
	i  *multiplex.ServiceCoreInternal
	db *db.DbClient
}

func NewReadDatabase(logger diag.Logger, dbClient *db.DbClient) *ReadDatabase {
	svc := &ReadDatabase{}
	svc.i = svc.InitServiceCore("ReadDatabase", logger, svc.coreProcessHook)
	svc.db = dbClient
	return svc
}

func (s *ReadDatabase) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	return &multiplex.HookState{Handled: true}
}
