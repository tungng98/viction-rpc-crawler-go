package svc

import (
	"sync"
	"time"

	"github.com/tforce-io/tf-golib/random/securerng"
)

type NetworkOptions struct {
	MaxRetries  int
	MaxRetryGap uint64
}

func (o *NetworkOptions) WaitRetryGap() {
	minRetryGap := uint64(20000000)
	if o.MaxRetryGap < minRetryGap {
		time.Sleep(time.Duration(minRetryGap))
	}
	duration := securerng.Uint64r(minRetryGap/2, o.MaxRetryGap)
	time.Sleep(time.Duration(duration))
}

type ProcessState int8

const (
	INIT_STATE ProcessState = iota
	EXIT_STATE
	SUCCESS_STATE
	ERROR_STATE
	RETRY_STATE
)

const (
	MAIN_CHAN_CAPACITY   = 256
	SECOND_CHAN_CAPACITY = 16
)

type BackgroundService interface {
	ServiceID() string
	// Controller() ServiceController
	SetWorker(workerCount uint16)
	WorkerCount() uint16
	Exec(command string, params ExecParams)
}

type ExecParams map[string]interface{}

func (p ExecParams) Get(key string, def interface{}) interface{} {
	if val, ok := p[key]; ok {
		return val
	}
	return def
}

func (p ExecParams) Set(key string, val interface{}) {
	p[key] = val
}

func (p ExecParams) Delete(key string) {
	delete(p, key)
}

func (p ExecParams) ExpectReturns() {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	p["returns"] = waitGroup
}

func (p ExecParams) WaitForReturns() {
	waitGroup := p["returns"].(*sync.WaitGroup)
	waitGroup.Wait()
}

type WorkerCounter struct {
	count uint16
	lock  sync.Mutex
}

func (c *WorkerCounter) Value() uint16 {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.count
}

func (c *WorkerCounter) ValueNoLock() uint16 {
	return c.count
}

func (c *WorkerCounter) Set(newCount uint16) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count = newCount
}

func (c *WorkerCounter) SetNoLock(newCount uint16) {
	c.count = newCount
}

func (c *WorkerCounter) Increase() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count++
}

func (c *WorkerCounter) IncreaseNoLock() {
	c.count++
}

func (c *WorkerCounter) Decrease() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count--
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
