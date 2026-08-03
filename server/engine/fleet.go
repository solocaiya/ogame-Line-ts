package engine

import "math"

// CalculateDistance computes the distance between two coordinates.
// Same position: 5
// Same system: 1000 + |pos1-pos2| * 5
// Same galaxy: 2700 + |sys1-sys2| * 95
// Different galaxy: 20000 + |gal1-gal2| * 20000
func CalculateDistance(from, to Coordinate) int {
	if from.Galaxy == to.Galaxy && from.System == to.System && from.Position == to.Position {
		return 5
	}
	if from.Galaxy == to.Galaxy && from.System == to.System {
		diff := from.Position - to.Position
		if diff < 0 {
			diff = -diff
		}
		return 1000 + diff*5
	}
	if from.Galaxy == to.Galaxy {
		diff := from.System - to.System
		if diff < 0 {
			diff = -diff
		}
		return 2700 + diff*95
	}
	diff := from.Galaxy - to.Galaxy
	if diff < 0 {
		diff = -diff
	}
	return 20000 + diff*20000
}

// CalculateFlightTime computes flight time in seconds.
// Formula: max(10, floor(distance * 50 / minSpeed))
func CalculateFlightTime(distance int, minSpeed int) int {
	if minSpeed <= 0 {
		return 10
	}
	time := int(math.Floor(float64(distance) * 50.0 / float64(minSpeed)))
	if time < 10 {
		time = 10
	}
	return time
}

// CalculateFuelConsumption computes fuel needed for a fleet mission.
func CalculateFuelConsumption(fleet map[string]int, distance, flightTime int) int64 {
	totalFuel := int64(0)
	for shipType, count := range fleet {
		def, ok := ShipDefs[shipType]
		if !ok {
			continue
		}
		// Fuel = fuelConsumption * distance / 10000 (simplified)
		fuel := int64(math.Ceil(float64(def.FuelConsumption) * float64(distance) / 10000.0))
		if fuel < 1 {
			fuel = 1
		}
		totalFuel += fuel * int64(count)
	}
	return totalFuel
}

// CalculateTotalCargoCapacity returns total cargo space for a fleet.
func CalculateTotalCargoCapacity(fleet map[string]int) int64 {
	total := int64(0)
	for shipType, count := range fleet {
		if def, ok := ShipDefs[shipType]; ok {
			total += int64(def.CargoCapacity) * int64(count)
		}
	}
	return total
}

// CalculateMinSpeed returns the slowest ship speed in a fleet.
func CalculateMinSpeed(fleet map[string]int) int {
	minSpeed := 0
	for shipType, count := range fleet {
		if count <= 0 {
			continue
		}
		def, ok := ShipDefs[shipType]
		if !ok {
			continue
		}
		if minSpeed == 0 || def.Speed < minSpeed {
			minSpeed = def.Speed
		}
	}
	return minSpeed
}
