package controller

import "backend-hotlines3/internal/feature/task/service"

type controller struct {
	service service.ServiceInterface
}

func NewController(service service.ServiceInterface) ControllerInterface {
	return &controller{service: service}
}

type ControllerInterface interface {
	V1ControllerInterface
}
