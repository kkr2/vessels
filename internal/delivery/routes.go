package delivery

import "github.com/labstack/echo/v4"

func MapVesselRoutes(vesselsGroup *echo.Group, h VesselsHandlers) {
	vesselsGroup.POST("", h.GetRoutesConsumtion())
}
