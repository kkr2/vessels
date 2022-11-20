package service

import (
	"context"
	"math"
	"sort"

	"github.com/kkr2/vessels/internal/domain"
	"github.com/kkr2/vessels/internal/logger"
	"github.com/kkr2/vessels/internal/repository/db"
	"github.com/kkr2/vessels/internal/repository/externalrpc"
)

// 1) Converts route to knot(speed)
// Given route , calculate average speed from point to point [] {src , dst , avgSpeed }

// 2) Get weather raport based on days.
// Given routes extract all days and get results . Client call + cache(mention).

// 3) Calculate aproximate consumption based on weather speed drought, refering to fuelTable
// Problem to be solved is if result we are searching is in between rows
// PRIORITY : Draught , Weather, Speed
// Optimisation: Get from db only what needed in between range
//(closest Drought existing in db)

type VesselService interface {
	GetRoutesConsumtion(ctx context.Context, imo int, drought float64, vesselRoutes []*domain.Route) ([]float64, error)
}

type vesselService struct {
	fuelRepo      db.VesselRepo
	weatherClient externalrpc.WeatherClient
	logger        logger.Logger
}

func NewVesselsService(fr db.VesselRepo, wc externalrpc.WeatherClient, log logger.Logger) VesselService {
	return &vesselService{
		fuelRepo:      fr,
		weatherClient: wc,
		logger:        log,
	}
}

type conResponse struct {
	index int
	res   float64
	err   error
}

func (vs *vesselService) GetRoutesConsumtion(ctx context.Context, imo int, drought float64, vesselRoutes []*domain.Route) ([]float64, error) {
	// TODO: Add validation
	allRouteFuelConsumtion := []float64{}

	fuelMaps, err := vs.fuelRepo.GetFuelMapWithClosestDrToTarget(ctx, imo, drought)
	if err != nil {
		return allRouteFuelConsumtion, err
	}

	for _, route := range vesselRoutes {
		r := route
		fm := fuelMaps
		routeConsumtion, err := vs.getRouteConsumtion(ctx, fm, r)
		if err != nil {
			return allRouteFuelConsumtion, err
		}
		allRouteFuelConsumtion = append(allRouteFuelConsumtion, routeConsumtion)
	}

	return allRouteFuelConsumtion, nil
}

// Calculates single route
func (vs *vesselService) getRouteConsumtion(ctx context.Context, fuelMap []*domain.FuelMap, vesselRoute *domain.Route) (float64, error) {
	//calculate avg speed point to point
	pointToPoints := vesselRoute.ConvertToP2P()
	//calculate avg weather point to point based on results that we got from api
	err := vs.calculateWeather(ctx, pointToPoints)
	if err != nil {
		return 0, err
	}
	//get most approximate consumtion , point to point based on draught , weather , speed
	vs.calculateConsumption(ctx, fuelMap, pointToPoints)

	//return total consumtion
	return calculateTotalConsumtion(pointToPoints), nil
}

// Updates pointToPoint data structure with weather information
func (vs *vesselService) calculateWeather(ctx context.Context, pointToPoints []*domain.PointToPoint) error {
	for _, ptp := range pointToPoints {
		ptp := ptp
		srcWeather, err := vs.weatherClient.GetWeatherForDay(ctx, ptp.Source.Date)
		if err != nil {
			return err
		}
		dstWeather, err := vs.weatherClient.GetWeatherForDay(ctx, ptp.Destination.Date)
		if err != nil {
			return err
		}
		ptp.AddWeatherInfo(srcWeather, dstWeather)

	}
	return nil
}

// Updates pointToPoint data structure with avg fuel consumption info
func (vs *vesselService) calculateConsumption(ctx context.Context, fuelMap []*domain.FuelMap, pointToPoints []*domain.PointToPoint) {
	for _, ptp := range pointToPoints {
		ptp := ptp
		avgConsumption := getClosestConsumtion(fuelMap, ptp.AvgSpeedInKnot, ptp.AvgWeatherInBeaufort)

		ptp.AddConsumtion(avgConsumption)

	}
}

// Gets closest sum from fuelmap provided by db, based on weather and speed (since closes to drought is provided by db)
func getClosestConsumtion(fuelMap []*domain.FuelMap, speed float64, weather float64) float64 {
	sort.Slice(fuelMap, func(i, j int) bool {
		if math.Abs(fuelMap[i].Weather-weather) != math.Abs(fuelMap[j].Weather-weather) {
			return math.Abs(fuelMap[i].Weather-weather) < math.Abs(fuelMap[j].Weather-weather)
		}
		return math.Abs(fuelMap[i].Speed-speed) < math.Abs(fuelMap[j].Speed-speed)

	})

	return fuelMap[0].Consumtion
}

// Helper function to add all exact consumtion from point to point data
func calculateTotalConsumtion(pointToPoints []*domain.PointToPoint) float64 {
	totalConsumption := 0.0
	for _, ptp := range pointToPoints {
		ptp := ptp
		totalConsumption += ptp.ExactConsumtion
	}
	return totalConsumption
}
