package svc

import (
	"sync"

	"github.com/rs/zerolog"
)

type ServiceControllerCommand struct {
	ServiceID string
	Command   string
	Params    map[string]interface{}
	ExecCount uint16
}

type ServiceController struct {
	i *ServiceControllerInternal
}

type ServiceControllerInternal struct {
	Services map[string]BackgroundService

	WorkerID    uint64
	WorkerCount uint16
	WorkerLock  sync.Mutex

	MainChan   chan *ServiceControllerCommand
	ExitChan   chan bool
	Background bool

	Logger *zerolog.Logger
}

func NewServiceController(logger *zerolog.Logger) ServiceController {
	svc := ServiceController{
		i: &ServiceControllerInternal{
			Services: map[string]BackgroundService{},
			MainChan: make(chan *ServiceControllerCommand),
			ExitChan: make(chan bool),
			Logger:   logger,
		},
	}
	return svc
}

func (s ServiceController) Exec(command string, params map[string]interface{}) {
	msg := &ServiceControllerCommand{
		ServiceID: "_",
		Command:   command,
		Params:    params,
		ExecCount: 0,
	}
	s.i.MainChan <- msg
}

func (s ServiceController) Run(background bool) {
	s.i.WorkerLock.Lock()
	if s.i.WorkerCount == 0 {
		if _, ok := s.i.Services["_"]; !ok {
			s.RegisterService("_", s)
		}
		s.i.WorkerCount = 1
		s.i.WorkerID++
		go s.process(s.i.WorkerID)
		s.i.Logger.Info().Msg("Controller started.")
	}

	if background {
		s.i.Background = true
		s.i.WorkerLock.Unlock()
		<-s.i.ExitChan
		s.i.Background = false
		s.i.Logger.Info().Msg("Controller exited.")
	} else {
		s.i.WorkerLock.Unlock()
	}
}

func (s ServiceController) SetWorker(workerCount uint16) {
}

func (s ServiceController) WorkerCount() uint16 {
	return s.i.WorkerCount
}

func (s ServiceController) Controller() ServiceController {
	return s
}

func (s *ServiceController) RegisterService(serviceID string, instance BackgroundService) {
	s.i.Services[serviceID] = instance
}

func (s *ServiceController) UnregisterService(serviceID string) {
	if _, ok := s.i.Services[serviceID]; !ok {
		return
	}
	if s.i.Services[serviceID].WorkerCount() == 0 {
		delete(s.i.Services, serviceID)
	}
}

func (s *ServiceController) process(workerID uint64) {
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Controller Process started.")
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.MainChan
		if msg.ServiceID == "_" {
			switch msg.Command {
			case "set_worker":
				serviceId := msg.Params["service_id"].(string)
				workerCount := msg.Params["worker_count"].(uint16)
				s.i.Services[serviceId].SetWorker(workerCount)
			case "exit":
				status = EXIT_STATE
			}
		} else {
			s.i.Services[msg.ServiceID].Exec(msg.Command, msg.Params)
			status = SUCCESS_STATE
		}
	}
	if s.i.Background {
		s.i.ExitChan <- true
	}
	s.i.WorkerCount--
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Controller Process exited.")
}
