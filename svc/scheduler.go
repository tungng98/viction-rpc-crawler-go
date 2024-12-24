package svc

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type JobMetadata struct {
	nextExecution int64

	IntervalMs int64
	Instance   BackgroundJob
	Params     map[string]interface{}
}

type ScheduleSvcCommand struct {
	Command string
}

type ScheduleSvc struct {
	i *ScheduleSvcInternal
}

type ScheduleSvcInternal struct {
	Jobs       map[string]*JobMetadata
	ExitSignal bool

	WorkerCount uint16
	WorkerLock  sync.Mutex
	WorkerID    uint64

	MainChan   chan *ScheduleSvcCommand
	ExitChan   chan bool
	Background bool

	IntervalMs int64
	Controller ServiceController
	Logger     *zerolog.Logger
}

func NewScheduleSvc(intervalMs int64, controller ServiceController, logger *zerolog.Logger) ScheduleSvc {
	svc := ScheduleSvc{
		i: &ScheduleSvcInternal{
			Jobs:       map[string]*JobMetadata{},
			MainChan:   make(chan *ScheduleSvcCommand),
			ExitChan:   make(chan bool),
			IntervalMs: intervalMs,
			Controller: controller,
			Logger:     logger,
		},
	}
	return svc
}

func (s ScheduleSvc) Exec(command string, params map[string]interface{}) {
}

func (s ScheduleSvc) Run(background bool) {
	s.i.WorkerLock.Lock()
	if s.i.WorkerCount == 0 {
		if s.i.IntervalMs < 250 {
			s.i.IntervalMs = 250
		}
		if s.i.IntervalMs > 60000 {
			s.i.IntervalMs = 60000
		}
		s.i.WorkerCount = 2
		s.i.WorkerID++
		go s.process(s.i.WorkerID)
		s.i.WorkerID++
		go s.processInterval(s.i.WorkerID)
		s.i.Logger.Info().Msg("Scheduler started.")
	}

	if background {
		s.i.Background = true
		s.i.WorkerLock.Unlock()
		<-s.i.ExitChan
		s.i.Logger.Info().Msg("Scheduler exited.")
		s.i.Background = false
	} else {
		s.i.WorkerLock.Unlock()
	}
}

func (s ScheduleSvc) SetWorker(workerCount uint16) {
}

func (s ScheduleSvc) WorkerCount() uint16 {
	return s.i.WorkerCount
}

func (s ScheduleSvc) Controller() ServiceController {
	return s.i.Controller
}

func (s *ScheduleSvc) process(workerID uint64) {
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler Process started.")
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.MainChan
		switch msg.Command {
		case "exit":
			status = EXIT_STATE
			s.i.ExitSignal = true
		}
	}
	s.i.WorkerCount--
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler Process exited.")
}

func (s *ScheduleSvc) processInterval(workerID uint64) {
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler ProcessInterval started.")
	for !s.i.ExitSignal {
		currentTimeMs := time.Now().UnixMilli()
		for jobID := range s.i.Jobs {
			job := s.i.Jobs[jobID]
			if job.nextExecution < currentTimeMs {
				if !job.Instance.IsExecuting() {
					job.Instance.Exec(job.Params)
				}
				job.nextExecution = currentTimeMs + job.IntervalMs
			}
		}
		time.Sleep(time.Duration(s.i.IntervalMs * int64(time.Millisecond)))
	}
	if s.i.Background {
		s.i.ExitChan <- true
	}
	s.i.WorkerCount--
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler Process exited.")
}
