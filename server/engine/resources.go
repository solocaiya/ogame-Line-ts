package engine

import "math"

// Resource production base values per hour
const (
	MetalMineBase      = 1500.0
	CrystalMineBase    = 1000.0
	DeuteriumBase      = 500.0
	DarkMatterBase     = 100.0
	SolarPlantBase     = 50.0
	FusionReactorBase  = 150.0
)

// Energy consumption base per level: {base, multiplier}
// Values match client src/config/gameConfig.ts exactly.
var buildingEnergyConsumption = map[string]struct {
	base       float64
	multiplier float64
}{
	"metalMine":           {10, 1.1},
	"crystalMine":         {10, 1.1},
	"deuteriumSynthesizer": {15, 1.1},
	"roboticsFactory":     {5, 1.1},
	"naniteFactory":       {20, 1.15},
	"shipyard":            {8, 1.1},
	"researchLab":         {12, 1.1},
	"terraformer":         {25, 1.12},
	"hangar":              {10, 1.1},
	"darkMatterCollector": {10, 1.1},
	"missileSilo":         {8, 1.1},
	"sensorPhalanx":       {15, 1.12},
	"jumpGate":            {50, 1.2},
}

// CalculateResourceProduction calculates hourly production for a planet.
// Returns production per hour and energy balance.
// Uses BINARY efficiency model: if energy production >= consumption, efficiency = 1; else 0.
// This matches the client's behavior in src/logic/resourceLogic.ts.
func CalculateResourceProduction(planet *PlanetState, gameSpeed int) (production Resources, energyProd, energyUsed int64, efficiency float64) {
	level := func(name string) int {
		if v, ok := planet.Buildings[name]; ok {
			return v
		}
		return 0
	}
	tech := func(name string) int {
		if v, ok := planet.Technologies[name]; ok {
			return v
		}
		return 0
	}

	speed := float64(gameSpeed)
	if speed <= 0 {
		speed = 1
	}

	// Metal mine: level * 1500 * 1.5^level * resourceBonus * metalTechBonus
	// mineralResearch gives +2% per level
	metalMineLvl := level("metalMine")
	metalProd := float64(metalMineLvl) * MetalMineBase * math.Pow(1.5, float64(metalMineLvl))
	metalProd *= (1 + 0.02*float64(tech("mineralResearch")))
	metalProd *= speed

	// Crystal mine: level * 1000 * 1.5^level
	// crystalResearch gives +2% per level
	crystalMineLvl := level("crystalMine")
	crystalProd := float64(crystalMineLvl) * CrystalMineBase * math.Pow(1.5, float64(crystalMineLvl))
	crystalProd *= (1 + 0.02*float64(tech("crystalResearch")))
	crystalProd *= speed

	// Deuterium: level * 500 * 1.5^level * tempBonus
	// fuelResearch gives +2% per level
	// Temperature bonus: 1.36 - 0.004 * maxTemp (lower temp = more deuterium)
	deuteriumLvl := level("deuteriumSynthesizer")
	deuteriumProd := float64(deuteriumLvl) * DeuteriumBase * math.Pow(1.5, float64(deuteriumLvl))
	deuteriumProd *= (1 + 0.02*float64(tech("fuelResearch")))
	deuteriumProd *= speed
	// Apply temperature bonus for deuterium
	deuteriumTempBonus := 1.36 - 0.004*float64(planet.MaxTemp)
	deuteriumProd *= deuteriumTempBonus

	// Dark matter: level * 100 * 1.5^level
	// NOT affected by energy efficiency, NOT affected by deposit efficiency
	darkMatterLvl := level("darkMatterCollector")
	darkMatterProd := float64(darkMatterLvl) * DarkMatterBase * math.Pow(1.5, float64(darkMatterLvl))
	darkMatterProd *= speed

	// Energy production
	// Solar plant: level * 50 * 1.1^level (NO energyTechnology bonus)
	solarLvl := level("solarPlant")
	solarEnergy := float64(solarLvl) * SolarPlantBase * math.Pow(1.1, float64(solarLvl))

	// Fusion reactor: level * 150 * 1.15^level (NO energyTechnology bonus)
	fusionLvl := level("fusionReactor")
	fusionEnergy := float64(fusionLvl) * FusionReactorBase * math.Pow(1.15, float64(fusionLvl))

	// Solar satellites: temperature-based output = count * floor((maxTemp + 160) / 6)
	solarSatCount := level("solarSatellite")
	solarSatPerUnit := math.Floor((float64(planet.MaxTemp) + 160.0) / 6.0)
	if solarSatPerUnit < 0 {
		solarSatPerUnit = 0
	}
	solarSatEnergy := float64(solarSatCount) * solarSatPerUnit

	totalEnergyProd := int64(solarEnergy + fusionEnergy + solarSatEnergy)

	// Energy consumption (NO energyTechnology reduction)
	totalEnergyUsed := int64(0)
	for building, params := range buildingEnergyConsumption {
		lvl := level(building)
		if lvl > 0 {
			consumption := float64(lvl) * params.base * math.Pow(params.multiplier, float64(lvl))
			totalEnergyUsed += int64(consumption)
		}
	}

	// BINARY efficiency: if production >= consumption, efficiency = 1; else 0
	if totalEnergyProd >= totalEnergyUsed {
		efficiency = 1.0
	} else {
		efficiency = 0.0
	}

	// Apply efficiency to mines (NOT to dark matter)
	metalProd *= efficiency
	crystalProd *= efficiency
	deuteriumProd *= efficiency

	production = Resources{
		Metal:      int64(metalProd),
		Crystal:    int64(crystalProd),
		Deuterium:  int64(deuteriumProd),
		DarkMatter: int64(darkMatterProd),
	}

	return production, totalEnergyProd, totalEnergyUsed, efficiency
}

// CalculateResourceCapacity returns storage capacity for standard resources.
// Base 10000 * 2^storageLevel * bonus
func CalculateResourceCapacity(storageLevel int, bonus float64) int64 {
	return int64(10000.0 * math.Pow(2, float64(storageLevel)) * bonus)
}

// CalculateDarkMatterCapacity returns storage capacity for dark matter.
// Base 1000 * 2^storageLevel * bonus (dark matter is rarer)
func CalculateDarkMatterCapacity(storageLevel int, bonus float64) int64 {
	return int64(1000.0 * math.Pow(2, float64(storageLevel)) * bonus)
}

// UpdatePlanetResources applies resource production since last update.
func UpdatePlanetResources(planet *PlanetState, now int64, gameSpeed int) {
	if planet.LastUpdate == 0 {
		planet.LastUpdate = now
		return
	}

	timeDiffMs := now - planet.LastUpdate
	if timeDiffMs <= 0 {
		return
	}

	production, energyProd, energyUsed, efficiency := CalculateResourceProduction(planet, gameSpeed)

	// Update energy state
	planet.EnergyProd = energyProd
	planet.EnergyUsed = energyUsed
	planet.Efficiency = efficiency
	planet.Production = production

	// Calculate resources gained (production is per hour, timeDiff is in ms)
	hoursElapsed := float64(timeDiffMs) / 3600000.0
	gained := Resources{
		Metal:      int64(float64(production.Metal) * hoursElapsed),
		Crystal:    int64(float64(production.Crystal) * hoursElapsed),
		Deuterium:  int64(float64(production.Deuterium) * hoursElapsed),
		DarkMatter: int64(float64(production.DarkMatter) * hoursElapsed),
	}

	// Add resources, capped at storage capacity
	planet.Resources.Metal = min(planet.Resources.Metal+gained.Metal, planet.StorageCap.Metal)
	planet.Resources.Crystal = min(planet.Resources.Crystal+gained.Crystal, planet.StorageCap.Crystal)
	planet.Resources.Deuterium = min(planet.Resources.Deuterium+gained.Deuterium, planet.StorageCap.Deuterium)
	planet.Resources.DarkMatter = min(planet.Resources.DarkMatter+gained.DarkMatter, planet.StorageCap.DarkMatter)

	planet.LastUpdate = now
}
