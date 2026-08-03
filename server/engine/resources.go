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
	SolarSatBase       = 0.0 // depends on temperature
)

// Energy consumption base per level
var buildingEnergyConsumption = map[string]struct {
	base     float64
	multiplier float64
}{
	"metalMine":          {30, 1.1},
	"crystalMine":        {20, 1.1},
	"deuteriumSynthesizer": {20, 1.1},
	"roboticsFactory":    {10, 1.1},
	"naniteFactory":      {20, 1.15},
	"shipyard":           {10, 1.1},
	"researchLab":        {10, 1.1},
	"terraformer":        {30, 1.15},
	"hangar":             {10, 1.1},
}

// CalculateResourceProduction calculates hourly production for a planet.
// Returns production per hour and energy balance.
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
	metalMineLvl := level("metalMine")
	metalProd := float64(metalMineLvl) * MetalMineBase * math.Pow(1.5, float64(metalMineLvl))
	metalProd *= (1 + 0.02*float64(tech("metalTechnology"))) // +2% per tech level
	metalProd *= speed

	// Crystal mine: level * 1000 * 1.5^level
	crystalMineLvl := level("crystalMine")
	crystalProd := float64(crystalMineLvl) * CrystalMineBase * math.Pow(1.5, float64(crystalMineLvl))
	crystalProd *= (1 + 0.02*float64(tech("crystalTechnology")))
	crystalProd *= speed

	// Deuterium: level * 500 * 1.5^level * tempBonus
	deuteriumLvl := level("deuteriumSynthesizer")
	deuteriumProd := float64(deuteriumLvl) * DeuteriumBase * math.Pow(1.5, float64(deuteriumLvl))
	deuteriumProd *= (1 + 0.02*float64(tech("deuteriumTechnology")))
	deuteriumProd *= speed

	// Dark matter: level * 100 * 1.5^level
	darkMatterLvl := level("darkMatterMine")
	darkMatterProd := float64(darkMatterLvl) * DarkMatterBase * math.Pow(1.5, float64(darkMatterLvl))
	darkMatterProd *= speed

	// Energy production
	// Solar plant: level * 50 * 1.1^level
	solarLvl := level("solarPlant")
	solarEnergy := float64(solarLvl) * SolarPlantBase * math.Pow(1.1, float64(solarLvl))
	solarEnergy *= (1 + 0.02*float64(tech("energyTechnology")))

	// Fusion reactor: level * 150 * 1.15^level
	fusionLvl := level("fusionReactor")
	fusionEnergy := float64(fusionLvl) * FusionReactorBase * math.Pow(1.15, float64(fusionLvl))
	fusionEnergy *= (1 + 0.02*float64(tech("energyTechnology")))

	// Solar satellites (simplified: assume average output)
	solarSatCount := level("solarSatellite")
	solarSatEnergy := float64(solarSatCount) * 25.0 // simplified average

	totalEnergyProd := int64(solarEnergy + fusionEnergy + solarSatEnergy)

	// Energy consumption
	totalEnergyUsed := int64(0)
	for building, params := range buildingEnergyConsumption {
		lvl := level(building)
		if lvl > 0 {
			consumption := float64(lvl) * params.base * math.Pow(params.multiplier, float64(lvl))
			// Energy tech reduces consumption by 1% per level
			consumption *= (1 - 0.01*float64(tech("energyTechnology")))
			totalEnergyUsed += int64(consumption)
		}
	}

	// Efficiency: production / consumption ratio
	if totalEnergyUsed > 0 && totalEnergyProd < totalEnergyUsed {
		efficiency = float64(totalEnergyProd) / float64(totalEnergyUsed)
	} else {
		efficiency = 1.0
	}

	// Apply efficiency to mines
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

// CalculateResourceCapacity returns storage capacity for each resource.
// Base 10000 * 2^storageLevel * bonus
func CalculateResourceCapacity(storageLevel int, bonus float64) int64 {
	return int64(10000.0 * math.Pow(2, float64(storageLevel)) * bonus)
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
