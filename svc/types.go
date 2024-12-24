package svc

type ProcessState int8

const (
	INIT_STATE ProcessState = iota
	EXIT_STATE
	SUCCESS_STATE
	ERROR_STATE
	RETRY_STATE
)

type BackgroundJob interface {
	Exec(params map[string]interface{})
	IsExecuting() bool
	Controller() ServiceController
}

type BackgroundService interface {
	Exec(command string, params map[string]interface{})
	Run(background bool)
	SetWorker(workerCount uint16)
	WorkerCount() uint16
	Controller() ServiceController
}
