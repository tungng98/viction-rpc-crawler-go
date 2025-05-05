package svc

import (
	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type ReadFileSystem struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

func NewReadFileSystem(logger diag.Logger) *ReadFileSystem {
	svc := &ReadFileSystem{}
	svc.i = svc.InitServiceCore("ReadFileSystem", logger, svc.coreProcessHook)
	return svc
}

func (s *ReadFileSystem) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	return &multiplex.HookState{Handled: true}
}
