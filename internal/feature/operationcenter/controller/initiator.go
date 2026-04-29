package controller

import "backend-hotlines3/internal/feature/operationcenter/service"

type Controller struct {
	service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{service: service}
}
