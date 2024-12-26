package svc

import (
	"github.com/rs/zerolog"
)

type ServiceControllerCommand struct {
	ServiceID string
	Command   string
	Params    ExecParams
	ExecCount uint16
}

type ServiceController struct {
	i *ServiceControllerInternal
}

type ServiceControllerInternal struct {
	Services map[string]BackgroundService

	WorkerMainCounter *WorkerCounter
	WorkerID          uint64

	MainChan   chan *ServiceControllerCommand
	ExitChan   chan bool
	Background bool

	Logger *zerolog.Logger
}

func NewServiceController(logger *zerolog.Logger) ServiceController {
	svc := ServiceController{
		i: &ServiceControllerInternal{
			Services: map[string]BackgroundService{},

			WorkerMainCounter: &WorkerCounter{},

			MainChan: make(chan *ServiceControllerCommand, MAIN_CHAN_CAPACITY),
			ExitChan: make(chan bool),

			Logger: logger,
		},
	}
	return svc
}

func (s ServiceController) Run(background bool) {
	s.i.WorkerMainCounter.Lock()
	if s.i.WorkerMainCounter.ValueNoLock() == 0 {
		if _, ok := s.i.Services[s.ServiceID()]; !ok {
			s.RegisterService(s)
		}
		s.i.WorkerID++
		go s.process(s.i.WorkerID)
		s.i.Logger.Info().Msg("Controller started.")
	}

	if background {
		s.i.Background = true
		s.i.WorkerMainCounter.Unlock()
		<-s.i.ExitChan
		s.i.Background = false
		s.i.Logger.Info().Msg("Controller exited.")
	} else {
		s.i.WorkerMainCounter.Unlock()
	}
}

func (s *ServiceController) RegisterService(instance BackgroundService) {
	s.i.Services[instance.ServiceID()] = instance
}

func (s *ServiceController) UnregisterService(serviceID string) {
	if _, ok := s.i.Services[serviceID]; !ok {
		return
	}
	if s.i.Services[serviceID].WorkerCount() == 0 {
		delete(s.i.Services, serviceID)
	}
}

func (s ServiceController) ServiceID() string {
	return "_"
}

func (s ServiceController) Controller() ServiceController {
	return s
}

func (s ServiceController) SetWorker(workerCount uint16) {
}

func (s ServiceController) WorkerCount() uint16 {
	return s.i.WorkerMainCounter.ValueNoLock()
}

func (s ServiceController) Exec(command string, params ExecParams) {
	msg := &ServiceControllerCommand{
		ServiceID: s.ServiceID(),
		Command:   command,
		Params:    params,
		ExecCount: 0,
	}
	s.i.MainChan <- msg
}

func (s ServiceController) ExecService(serviceID, command string, params ExecParams) {
	msg := &ServiceControllerCommand{
		ServiceID: serviceID,
		Command:   command,
		Params:    params,
		ExecCount: 0,
	}
	s.i.MainChan <- msg
}

func (s *ServiceController) process(workerID uint64) {
	s.i.WorkerMainCounter.Increase()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Controller Process started.")
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.MainChan
		if msg.ServiceID == s.ServiceID() {
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
	s.i.WorkerMainCounter.Decrease()
	s.i.Logger.Info().Uint64("WorkerID", workerID).Msg("Controller Process exited.")
}
