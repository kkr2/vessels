package delivery

import (
	"github.com/kkr2/vessels/internal/domain"
)

type GetRoutesConsumptionRequest struct {
	Imo     int             `json:"imo" validate:"required"`
	Draught float64         `json:"draught" validate:"required"`
	Routes  []*domain.Route `json:"routes" validate:"required"`
}
