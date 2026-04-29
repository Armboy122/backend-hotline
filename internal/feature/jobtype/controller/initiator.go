package controller

import "backend-hotlines3/internal/feature/jobtype/service"

type Controller struct {
	service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{service: service}
}
