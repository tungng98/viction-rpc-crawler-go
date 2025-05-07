package svc

import (
	"fmt"
	"math/big"
	"path/filepath"
	"viction-rpc-crawler-go/filesystem"

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
		blockDatas := msg.GetParam("blocks", []*GetBlockResult{}).([]*GetBlockResult)
		rootDir := msg.GetParam("root", "").(string)
		if rootDir == "" {
			s.i.Logger.Warnf("%s#%d: RootDir is empty. No files will be written.", s.ServiceID(), workerID)
			msg.Return(nil)
			break
		}
		outputDir := filepath.Join(rootDir, "getBlockByNumber")
		for _, blockData := range blockDatas {
			midDirs := GetNumberedDir(blockData.Number)
			blockFile := filepath.Join(outputDir, midDirs[0], midDirs[1], blockData.Number.String()+".json")
			err := filesystem.WriteFile(blockFile, []byte(blockData.RawData))
			if err != nil {
				s.i.Logger.Errorf(err, "%s#%d: Failed to write block file #%d.", s.ServiceID(), workerID, blockData.Number.Uint64())
				continue
			}
		}
		msg.Return(true)
	case "debug_traceBlockByNumber":
		blockTraces := msg.GetParam("block_traces", []*TraceBlockResult{}).([]*TraceBlockResult)
		rootDir := msg.GetParam("root", "").(string)
		if rootDir == "" {
			s.i.Logger.Warnf("%s#%d: RootDir is empty. No files will be written.", s.ServiceID(), workerID)
			msg.Return(nil)
			break
		}
		outputDir := filepath.Join(rootDir, "traceBlockByNumber")
		for _, blockTrace := range blockTraces {
			midDirs := GetNumberedDir(blockTrace.Number)
			blockTraceFile := filepath.Join(outputDir, midDirs[0], midDirs[1], blockTrace.Number.String()+".json")
			err := filesystem.WriteFile(blockTraceFile, []byte(blockTrace.RawData))
			if err != nil {
				s.i.Logger.Errorf(err, "%s#%d: Failed to write block trace file #%d.", s.ServiceID(), workerID, blockTrace.Number.Uint64())
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

func GetNumberedDir(number *big.Int) []string {
	paddedNumber := fmt.Sprintf("%09d", number.Uint64())
	length := len(paddedNumber)
	firstLevel := paddedNumber[length-9 : length-6]
	secondLevel := paddedNumber[length-6 : length-3]
	thirdLevel := paddedNumber[length-3 : length]
	return []string{firstLevel, secondLevel, thirdLevel}
}
