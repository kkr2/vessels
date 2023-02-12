package delivery

import (
	"net/http"

	"github.com/kkr2/vessels/internal/config"
	"github.com/kkr2/vessels/internal/logger"
	"github.com/kkr2/vessels/internal/service"
	"github.com/labstack/echo/v4"
)

type VesselsHandlers interface {
	GetRoutesConsumtion() echo.HandlerFunc
}

type vesselsHandlers struct {
	cfg    *config.Config
	vs     service.VesselService
	logger logger.Logger
}

// NewNewsHandlers News handlers constructor
func NewVesselsHandlers(cfg *config.Config, vs service.VesselService, logger logger.Logger) VesselsHandlers {
	return &vesselsHandlers{cfg: cfg, vs: vs, logger: logger}
}

func (h vesselsHandlers) GetRoutesConsumtion() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := GetRequestCtx(c)

		req := &GetRoutesConsumptionRequest{}

		if err := SanitizeRequest(c, req); err != nil {
			return ErrResponseWithLog(c, h.logger, err)
		}

		rowRes, err := h.vs.GetRoutesConsumtion(ctx, req.Imo, req.Draught, req.Routes)
		if err != nil {
			LogResponseError(c, h.logger, err)
			return c.JSON(ErrorResponse(err))
		}
		return c.JSON(http.StatusOK, NewResponseView(rowRes))
	}
}
