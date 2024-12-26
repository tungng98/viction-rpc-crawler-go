package svc

import (
	"time"

	"github.com/rs/zerolog"
)

type JobMetadata struct {
	nextExecution int64

	IntervalMs int64
	ServiceID  string
	Command    string
	Params     ExecParams
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

	WorkerMainCounter *WorkerCounter
	WorkerMainCount   uint16
	WorkerID          uint64

	MainChan chan *ScheduleSvcCommand

	IntervalMs int64
	Controller ServiceController
	Logger     *zerolog.Logger
}

func NewScheduleSvc(intervalMs int64, controller ServiceController, logger *zerolog.Logger) ScheduleSvc {
	svc := ScheduleSvc{
		i: &ScheduleSvcInternal{
			Jobs: map[string]*JobMetadata{},

			WorkerMainCounter: &WorkerCounter{},

			MainChan: make(chan *ScheduleSvcCommand, MAIN_CHAN_CAPACITY),

			IntervalMs: intervalMs,
			Controller: controller,
			Logger:     logger,
		},
	}
	return svc
}

func (s ScheduleSvc) AddJob(jobID string, intervalMs int64, serviceID string, command string, params ExecParams) {
	s.i.Jobs[jobID] = &JobMetadata{
		IntervalMs: intervalMs,
		ServiceID:  serviceID,
		Command:    command,
		Params:     params,
	}
}

func (s ScheduleSvc) ServiceID() string {
	return "Scheduler"
}

func (s ScheduleSvc) Controller() ServiceController {
	return s.i.Controller
}

func (s ScheduleSvc) SetWorker(workerCount uint16) {
	s.i.WorkerMainCounter.Lock()
	defer s.i.WorkerMainCounter.Unlock()
	if s.i.WorkerMainCounter.ValueNoLock() != s.i.WorkerMainCount {
		return
	}
	if workerCount > 0 && s.i.WorkerMainCounter.ValueNoLock() == 0 {
		if s.i.IntervalMs < 250 {
			s.i.IntervalMs = 250
		}
		if s.i.IntervalMs > 60000 {
			s.i.IntervalMs = 60000
		}
		s.i.WorkerMainCount = 2
		s.i.WorkerID++
		go s.process(s.i.WorkerID)
		s.i.WorkerID++
		go s.processInterval(s.i.WorkerID)
		s.i.Logger.Info().Msg("Scheduler started.")
	}
	if workerCount == 0 && s.i.WorkerMainCounter.ValueNoLock() > 0 {
		s.i.WorkerMainCount = 0
		cmd := &ScheduleSvcCommand{
			Command: "exit",
		}
		s.i.MainChan <- cmd
	}
}

func (s ScheduleSvc) WorkerCount() uint16 {
	return s.i.WorkerMainCounter.ValueNoLock()
}

func (s ScheduleSvc) Exec(command string, params ExecParams) {
	if command == "exit" {
		return
	}
}

func (s *ScheduleSvc) process(workerID uint64) {
	s.i.WorkerMainCounter.Increase()
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
	s.i.WorkerMainCounter.Decrease()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler Process exited.")
}

func (s *ScheduleSvc) processInterval(workerID uint64) {
	s.i.WorkerMainCounter.Increase()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler ProcessInterval started.")
	for !s.i.ExitSignal {
		currentTimeMs := time.Now().UnixMilli()
		for jobID := range s.i.Jobs {
			job := s.i.Jobs[jobID]
			if job.nextExecution < currentTimeMs {
				s.i.Controller.ExecService(job.ServiceID, job.Command, job.Params)
				job.nextExecution = currentTimeMs + job.IntervalMs
			}
		}
		time.Sleep(time.Duration(s.i.IntervalMs * int64(time.Millisecond)))
	}
	s.i.WorkerMainCounter.Decrease()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Scheduler Process exited.")
}
