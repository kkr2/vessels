package db


const (
	findClosestFuelConsumtion = ` SELECT * FROM fuel f
									WHERE f.imo = $1 
									ORDER BY ABS(draught - $2) , ABS(beaufort  - $3) , ABS(speed  - $4)
									LIMIT 1`

	allFuelMapsWithClosestDr = ` select *
								from fuel f 
								where f.imo = $1 and draught = (
									SELECT f.draught 
									FROM fuel f
									WHERE f.imo = $1 
									ORDER BY ABS(draught - $2)
									limit 1
								)`
)
