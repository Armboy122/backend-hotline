package controller

import "backend-hotlines3/internal/feature/jobdetail/service"

type Controller struct {
	service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{service: service}
}
