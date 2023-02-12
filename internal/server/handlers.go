package server

import (
	"net/http"

	"github.com/kkr2/vessels/internal/delivery"
	"github.com/kkr2/vessels/internal/repository/db"
	"github.com/kkr2/vessels/internal/repository/externalrpc"
	"github.com/kkr2/vessels/internal/service"
	"github.com/labstack/echo/v4"
)

func (s *Server) MapHandlers(e *echo.Echo) error {

	// Init repositories
	vRepo := db.NewVesselsRepository(s.db, s.logger)
	vClient := externalrpc.NewWeatherClient(s.cfg, s.logger)

	// Init useCases
	vService := service.NewVesselsService(vRepo, vClient, s.logger)

	// Init handlers
	vHandler := delivery.NewVesselsHandlers(s.cfg, vService, s.logger)

	v1 := e.Group("/api/v1")

	health := v1.Group("/health")
	vesselGroup := v1.Group("/vessels")

	delivery.MapVesselRoutes(vesselGroup, vHandler)

	health.GET("", func(c echo.Context) error {
		s.logger.Infof("Health check RequestID: %s", c.Response().Header().Get(echo.HeaderXRequestID))
		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	})

	return nil
}
