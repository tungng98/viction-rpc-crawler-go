package svc

import (
	"sync"

	"github.com/tforce-io/tf-golib/multiplex"
)

type ReturnParams struct {
	Signal *sync.WaitGroup
	Result interface{}
}

func ExpectReturns(params multiplex.ExecParams) {
	signal := new(sync.WaitGroup)
	signal.Add(1)
	params["returns"] = &ReturnParams{
		Signal: signal,
	}
}

func ExpectReturnsWithCustomSignal(params multiplex.ExecParams, signal *sync.WaitGroup) {
	params["returns"] = &ReturnParams{
		Signal: signal,
	}
}

func WaitForReturns(msg *multiplex.ServiceMessage) interface{} {
	if msg.Params["returns"] != nil {
		returns := msg.Params["returns"].(*ReturnParams)
		returns.Signal.Wait()
		return returns.Result
	}
	return nil
}

func Return(msg *multiplex.ServiceMessage, result interface{}) {
	if msg.Params["returns"] != nil {
		returns := msg.Params["returns"].(*ReturnParams)
		returns.Result = result
	}
}

func ReturnSync(msg *multiplex.ServiceMessage, result interface{}) {
	if msg.Params["returns"] != nil {
		returns := msg.Params["returns"].(*ReturnParams)
		returns.Result = result
		returns.Signal.Done()
	}
}
