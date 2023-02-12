package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/kkr2/vessels/internal/domain"
	"github.com/kkr2/vessels/internal/errors"
	"github.com/kkr2/vessels/internal/logger"
)

// Repository Interface
type VesselRepo interface {
	GetClosestConsumtion(
		ctx context.Context,
		imo int,
		draught float64,
		speed float64,
		beaufort float64,
	) (float64, error)

	GetFuelMapWithClosestDrToTarget(
		ctx context.Context,
		imo int,
		draught float64,
	) ([]*domain.FuelMap, error)
}

// Vessels Repository
type vesselRepo struct {
	db  *sqlx.DB
	log logger.Logger
}

// Vessels repository constructor
func NewVesselsRepository(db *sqlx.DB, log logger.Logger) VesselRepo {
	return &vesselRepo{db: db, log: log}
}

func (vr *vesselRepo) GetClosestConsumtion(
	ctx context.Context,
	imo int,
	draught float64,
	speed float64,
	beaufort float64,
) (float64, error) {

	operation := errors.Op("db.vesselsRepository.GetClosestConsumption")

	fRow := &domain.FuelMap{}

	if err := vr.db.QueryRowxContext(
		ctx, findClosestFuelConsumtion,
		imo,
		draught,
		beaufort,
		speed,
	).StructScan(fRow); err != nil {
		return 0, errors.E(operation, errors.KindInternal, err)
	}

	return fRow.Consumtion, nil
}
func (vr *vesselRepo) GetFuelMapWithClosestDrToTarget(
	ctx context.Context,
	imo int,
	draught float64,
) ([]*domain.FuelMap, error) {

	operation := errors.Op("db.vesselsRepository.GetFuelMapWithClosestDrToTarget")

	rows, err := vr.db.QueryxContext(ctx, allFuelMapsWithClosestDr, imo, draught)

	if err != nil {
		return nil, errors.E(operation, errors.KindInternal, err)
	}
	defer rows.Close()

	fuelList := make([]*domain.FuelMap, 0)
	for rows.Next() {
		fuelMap := &domain.FuelMap{}
		if err = rows.StructScan(fuelMap); err != nil {
			return nil, errors.E(operation, errors.KindInternal, err)
		}
		fuelList = append(fuelList, fuelMap)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.E(operation, errors.KindInternal, err)
	}

	return fuelList, nil

}
