package controller

import "backend-hotlines3/internal/feature/auth/service"

type Controller struct {
	service *service.Service
}

func NewController(svc *service.Service) *Controller {
	return &Controller{service: svc}
}
