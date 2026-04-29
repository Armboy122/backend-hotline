package monthlyplan

import (
	v1 "backend-hotlines3/internal/handlers/v1"
)

// Controller is the module-first entrypoint for MonthlyPlan routes.
// It currently wraps the existing v1 handler while the module is being migrated.
type Controller struct {
	*v1.MonthlyPlanHandler
}

// NewController creates a Controller that delegates to the existing v1 handler.
func NewController(handler *v1.MonthlyPlanHandler) *Controller {
	return &Controller{MonthlyPlanHandler: handler}
}
