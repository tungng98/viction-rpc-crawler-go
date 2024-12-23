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
	services    map[string]BackgroundService
	workerCount uint16
	workerLock  sync.Mutex
	mainChan    chan *ServiceControllerCommand
	exitChan    chan bool

	Logger *zerolog.Logger
}

func NewServiceController(logger *zerolog.Logger) ServiceController {
	svc := ServiceController{
		i: &ServiceControllerInternal{
			services: map[string]BackgroundService{},
			mainChan: make(chan *ServiceControllerCommand),
			exitChan: make(chan bool),
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
	s.i.mainChan <- msg
}

func (s ServiceController) Run() {
	s.i.workerLock.Lock()
	if _, ok := s.i.services["_"]; !ok {
		s.RegisterService("_", s)
	}
	if s.i.workerCount == 0 {
		s.i.workerCount = 1
		go s.process()
	}
	s.i.workerLock.Unlock()
	s.i.Logger.Info().Msg("Service Controller started")

	<-s.i.exitChan
}

func (s ServiceController) Start(workerCount uint16) {
}

func (s ServiceController) Stop(workerCount uint16) {
}

func (s ServiceController) WorkerCount() uint16 {
	return s.i.workerCount
}

func (s *ServiceController) RegisterService(serviceID string, instance BackgroundService) {
	s.i.services[serviceID] = instance
}

func (s *ServiceController) UnregisterService(serviceID string) {
	if _, ok := s.i.services[serviceID]; !ok {
		return
	}
	if s.i.services[serviceID].WorkerCount() == 0 {
		delete(s.i.services, serviceID)
	}
}

func (s *ServiceController) process() {
	status := INIT_STATE
	for status != EXIT_STATE {
		msg := <-s.i.mainChan
		if msg.ServiceID == "_" {
			switch msg.Command {
			case "start":
				serviceId := msg.Params["service_id"].(string)
				workerCount := msg.Params["worker_count"].(uint16)
				s.i.services[serviceId].Start(workerCount)
			case "stop":
				serviceId := msg.Params["service_id"].(string)
				workerCount := msg.Params["worker_count"].(uint16)
				s.i.services[serviceId].Stop(workerCount)
			case "exit":
				status = EXIT_STATE
			}
		} else {
			s.i.services[msg.ServiceID].Exec(msg.Command, msg.Params)
			status = SUCCESS_STATE
		}
	}
	s.i.exitChan <- true
}
