package engine

import "math"

// CalculateBuildingCost returns the cost to upgrade a building to targetLevel.
// Formula: floor(baseCost * costMultiplier^(targetLevel-1))
func CalculateBuildingCost(baseCost CostEntry, costMultiplier float64, targetLevel int) CostEntry {
	multiplier := math.Pow(costMultiplier, float64(targetLevel-1))
	return CostEntry{
		Metal:     int64(math.Floor(float64(baseCost.Metal) * multiplier)),
		Crystal:   int64(math.Floor(float64(baseCost.Crystal) * multiplier)),
		Deuterium: int64(math.Floor(float64(baseCost.Deuterium) * multiplier)),
	}
}

// CalculateBuildingTime returns build time in seconds.
// Formula: floor(baseTime / (1 + roboticsLevel + naniteLevel*2) * (1 - speedBonus/100))
func CalculateBuildingTime(baseTime int, roboticsLevel, naniteLevel int, speedBonus float64) int {
	divisor := 1.0 + float64(roboticsLevel) + float64(naniteLevel)*2.0
	time := float64(baseTime) / divisor
	time *= (1.0 - speedBonus/100.0)
	result := int(math.Floor(time))
	if result < 1 {
		result = 1
	}
	return result
}

// CalculateShipBuildTime returns the time to build one ship.
func CalculateShipBuildTime(baseTime int, shipyardLevel, naniteLevel int) int {
	divisor := 1.0 + float64(shipyardLevel)*0.5 + float64(naniteLevel)*2.0
	time := float64(baseTime) / divisor
	result := int(math.Floor(time))
	if result < 1 {
		result = 1
	}
	return result
}

// CalculateDefenseBuildTime returns the time to build one defense unit.
func CalculateDefenseBuildTime(baseTime int, shipyardLevel, naniteLevel int) int {
	// Same as ship build time
	return CalculateShipBuildTime(baseTime, shipyardLevel, naniteLevel)
}

// CalculateUsedSpace returns total space used by buildings on a planet.
func CalculateUsedSpace(buildings map[string]int) int {
	total := 0
	for buildingType, level := range buildings {
		if def, ok := BuildingDefs[buildingType]; ok {
			total += def.SpaceUsage * level
		}
	}
	return total
}
