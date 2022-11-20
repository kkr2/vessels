package delivery

type RouteConsumptionResponse struct {
	ConsumtionInMetricTons float64 `json:"ConsumtionInMetricTons"`
	ConsumptionInCO2       float64 `json:"ConsumptionInCO2"`
}

func NewResponseView(consumtions []float64) []RouteConsumptionResponse {
	allRoutesConsumption := []RouteConsumptionResponse{}

	for _, consumtion := range consumtions {
		r := RouteConsumptionResponse{
			ConsumtionInMetricTons: consumtion,
			ConsumptionInCO2:       consumtion * float64(3.114),
		}
		allRoutesConsumption = append(allRoutesConsumption, r)
	}
	return allRoutesConsumption
}
