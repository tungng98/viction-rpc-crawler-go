package svc

import (
	"encoding/json"
	"path/filepath"
	"viction-rpc-crawler-go/filesystem"
	"viction-rpc-crawler-go/rpc"

	"github.com/tforce-io/tf-golib/diag"
	"github.com/tforce-io/tf-golib/multiplex"
)

type WriteFileSystem struct {
	multiplex.ServiceCore
	i *multiplex.ServiceCoreInternal
}

func NewWriteFileSystem(logger diag.Logger) *WriteFileSystem {
	svc := &WriteFileSystem{}
	svc.i = svc.InitServiceCore("WriteFileSystem", logger, svc.coreProcessHook)
	return svc
}

func (s *WriteFileSystem) coreProcessHook(workerID uint64, msg *multiplex.ServiceMessage) *multiplex.HookState {
	switch msg.Command {
	case "eth_getBlockByNumber":
		blockJSONs := msg.GetParam("blocks", []string{}).([]string)
		rootDir := msg.GetParam("root", "").(string)
		if rootDir == "" {
			s.i.Logger.Warnf("%s#%d: RootDir is empty. No files will be written.", s.ServiceID(), workerID)
			break
		}
		outputDir := filepath.Join(rootDir, "getBlockByNumber")
		for _, blockJSON := range blockJSONs {
			var block *rpc.Block
			err := json.Unmarshal([]byte(blockJSON), &block)
			if err != nil {
				s.i.Logger.Errorf(err, "%s#%d: Failed to write block file.", s.ServiceID(), workerID)
				continue
			}
			blockFile := filepath.Join(outputDir, block.Number.BigInt().String()+".json")
			err = filesystem.WriteFile(blockFile, []byte(blockJSON))
			if err != nil {
				s.i.Logger.Errorf(err, "%s#%d: Failed to write block file #%d.", s.ServiceID(), workerID, block.Number.BigInt().Uint64())
				continue
			}
		}
		msg.Return(true)
	default:
		s.i.Logger.Warnf("%s#%d: Unknown command %s.", s.i.ServiceID, workerID, msg.Command)
		msg.Return(nil)
	}
	return &multiplex.HookState{Handled: true}
}
