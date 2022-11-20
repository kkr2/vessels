package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type FuelMap struct {
	ID         uuid.UUID `db:"id"`
	VesselId   int       `db:"imo"`
	Draught    float64   `db:"draught"`
	Weather    float64   `db:"beaufort"`
	Speed      float64   `db:"speed"`
	Consumtion float64   `db:"consumption"`
}

type Route []RouteData

type RouteData struct {
	Date      time.Time `json:"date"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
}

type PointToPoint struct {
	Source               RouteData
	Destination          RouteData
	TimeDiffInMins       float64
	AvgSpeedInKnot       float64
	AvgWeatherInBeaufort float64
	AvgDailyConsumtion   float64
	ExactConsumtion      float64
}

// Coverts given data points to pointToPoint data
func (route *Route) ConvertToP2P() []*PointToPoint {
	allRoutePoints := []*PointToPoint{}

	// TODO: sort based on time

	for i := 1; i < len(*route); i++ {

		timeDiff := (*route)[i].Date.Sub((*route)[i-1].Date)

		newP2P := &PointToPoint{
			Source:         (*route)[i-1],
			Destination:    (*route)[i],
			TimeDiffInMins: timeDiff.Minutes(),
		}
		newP2P.calculateAvgSpeed()
		allRoutePoints = append(allRoutePoints, newP2P)
	}

	return allRoutePoints
}

// Calculates avg speed in kn given 2 locations with respective time
func (point *PointToPoint) calculateAvgSpeed() {
	distanceInNM := distanceinN(point.Source.Latitude, point.Source.Longitude, point.Destination.Latitude, point.Source.Longitude)

	point.AvgSpeedInKnot = (distanceInNM * 60.0) / point.TimeDiffInMins
}

// Calculates distance in N given locations
func distanceinN(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	return dist * 0.8684
}

// FIX: This function is supposing that any 2 points have consecutive days
// Updates object with weather info required
func (point *PointToPoint) AddWeatherInfo(srcPointBeaufort, destPointBeaufort float64) {
	if srcPointBeaufort == destPointBeaufort {
		point.AvgWeatherInBeaufort = srcPointBeaufort
		return
	}
	//TODO: Make this more accurate
	point.AvgWeatherInBeaufort = (srcPointBeaufort + destPointBeaufort) / 2.0

}

// Updates object with consumption info required
func (point *PointToPoint) AddConsumtion(avgConsumption float64) {
	// update avg consumption
	point.AvgDailyConsumtion = avgConsumption
	// calculate fraction of the day it takes vessel to go from point to point
	dayFraction := point.TimeDiffInMins / float64(1440)

	point.ExactConsumtion = avgConsumption * dayFraction

}
