package svc

import "sync"

type ProcessState int8

const (
	INIT_STATE ProcessState = iota
	EXIT_STATE
	SUCCESS_STATE
	ERROR_STATE
	RETRY_STATE
)

type BackgroundService interface {
	ServiceID() string
	Controller() ServiceController
	SetWorker(workerCount uint16)
	WorkerCount() uint16
	Exec(command string, params ExecParams)
}

type ExecParams map[string]interface{}

func (p *ExecParams) ExpectReturns() {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	(*p)["returns"] = waitGroup
}

func (p *ExecParams) WaitForReturns() {
	waitGroup := (*p)["returns"].(*sync.WaitGroup)
	waitGroup.Wait()
}

type WorkerCounter struct {
	count uint16
	lock  sync.Mutex
}

func (c *WorkerCounter) Value() uint16 {
	c.lock.Lock()
	count := c.count
	c.lock.Unlock()
	return count
}

func (c *WorkerCounter) ValueNoLock() uint16 {
	return c.count
}

func (c *WorkerCounter) Set(newCount uint16) {
	c.lock.Lock()
	c.count = newCount
	c.lock.Unlock()
}

func (c *WorkerCounter) SetNoLock(newCount uint16) {
	c.count = newCount
}

func (c *WorkerCounter) Increase() {
	c.lock.Lock()
	c.count++
	c.lock.Unlock()
}

func (c *WorkerCounter) IncreaseNoLock() {
	c.count++
}

func (c *WorkerCounter) Decrease() {
	c.lock.Lock()
	c.count--
	c.lock.Unlock()
}

func (c *WorkerCounter) DecreaseNoLock() {
	c.count--
}

func (c *WorkerCounter) Lock() {
	c.lock.Lock()
}

func (c *WorkerCounter) Unlock() {
	c.lock.Unlock()
}
