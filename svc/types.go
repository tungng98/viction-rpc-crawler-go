package svc

type ProcessState int8

const (
	INIT_STATE ProcessState = iota
	EXIT_STATE
	SUCCESS_STATE
	ERROR_STATE
	RETRY_STATE
)

type BackgroundService interface {
	Exec(command string, params map[string]interface{})
	Start(workerCount uint16)
	Stop(workerCount uint16)
	WorkerCount() uint16
}
