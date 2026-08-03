package engine

// BuildingDefs contains all building definitions.
var BuildingDefs = map[string]BuildingDef{
	"metalMine":              {BaseCost: CostEntry{Metal: 60, Crystal: 15}, BaseTime: 15, CostMultiplier: 1.5, SpaceUsage: 1},
	"crystalMine":            {BaseCost: CostEntry{Metal: 48, Crystal: 24}, BaseTime: 15, CostMultiplier: 1.6, SpaceUsage: 1},
	"deuteriumSynthesizer":   {BaseCost: CostEntry{Metal: 225, Crystal: 75}, BaseTime: 30, CostMultiplier: 1.5, SpaceUsage: 1},
	"solarPlant":             {BaseCost: CostEntry{Metal: 75, Crystal: 30}, BaseTime: 20, CostMultiplier: 1.5, SpaceUsage: 1},
	"fusionReactor":          {BaseCost: CostEntry{Metal: 900, Crystal: 360, Deuterium: 180}, BaseTime: 100, CostMultiplier: 1.8, SpaceUsage: 1},
	"roboticsFactory":        {BaseCost: CostEntry{Metal: 400, Crystal: 120, Deuterium: 200}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"naniteFactory":          {BaseCost: CostEntry{Metal: 1000000, Crystal: 500000, Deuterium: 100000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 1},
	"shipyard":               {BaseCost: CostEntry{Metal: 400, Crystal: 200, Deuterium: 100}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"hangar":                 {BaseCost: CostEntry{Metal: 2000, Crystal: 2000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 1},
	"researchLab":            {BaseCost: CostEntry{Metal: 200, Crystal: 400, Deuterium: 200}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"metalStorage":           {BaseCost: CostEntry{Metal: 1000}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"crystalStorage":         {BaseCost: CostEntry{Metal: 1000, Crystal: 500}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"deuteriumTank":          {BaseCost: CostEntry{Metal: 1000, Crystal: 1000}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 1},
	"darkMatterCollector":    {BaseCost: CostEntry{Metal: 5000, Crystal: 2500, Deuterium: 1000}, BaseTime: 120, CostMultiplier: 1.8, SpaceUsage: 1},
	"darkMatterTank":         {BaseCost: CostEntry{Metal: 10000, Crystal: 5000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 1},
	"missileSilo":            {BaseCost: CostEntry{Metal: 20000, Crystal: 20000, Deuterium: 1000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 1},
	"terraformer":            {BaseCost: CostEntry{Crystal: 1000000, Deuterium: 500000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 0},
	"lunarBase":              {BaseCost: CostEntry{Metal: 20000, Crystal: 40000, Deuterium: 20000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 0},
	"sensorPhalanx":          {BaseCost: CostEntry{Metal: 20000, Crystal: 40000, Deuterium: 20000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 1},
	"jumpGate":               {BaseCost: CostEntry{Metal: 2000000, Crystal: 4000000, Deuterium: 2000000}, BaseTime: 600, CostMultiplier: 2.0, SpaceUsage: 1},
	"planetDestroyerFactory": {BaseCost: CostEntry{Metal: 5000000, Crystal: 4000000, Deuterium: 2000000}, BaseTime: 600, CostMultiplier: 2.0, SpaceUsage: 1},
	"geoResearchStation":     {BaseCost: CostEntry{Metal: 500000, Crystal: 300000, Deuterium: 200000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 1},
	"deepDrillingFacility":   {BaseCost: CostEntry{Metal: 1000000, Crystal: 500000, Deuterium: 300000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 1},
	"university":             {BaseCost: CostEntry{Metal: 2000000, Crystal: 1000000, Deuterium: 500000}, BaseTime: 600, CostMultiplier: 2.0, SpaceUsage: 1},
}

// ShipDefs contains all ship definitions with combat stats.
var ShipDefs = map[string]ShipDef{
	"lightFighter": {
		Cost: CostEntry{Metal: 3000, Crystal: 1000}, BuildTime: 30,
		CargoCapacity: 50, Attack: 50, Shield: 10, Armor: 40,
		Speed: 12500, FuelConsumption: 20, StorageUsage: 1,
		RapidFire: map[string]int{"espionageProbe": 5},
	},
	"heavyFighter": {
		Cost: CostEntry{Metal: 6000, Crystal: 4000}, BuildTime: 60,
		CargoCapacity: 100, Attack: 150, Shield: 25, Armor: 75,
		Speed: 10000, FuelConsumption: 75, StorageUsage: 2,
		RapidFire: map[string]int{"espionageProbe": 5, "smallCargo": 3},
	},
	"cruiser": {
		Cost: CostEntry{Metal: 20000, Crystal: 7000, Deuterium: 2000}, BuildTime: 120,
		CargoCapacity: 800, Attack: 400, Shield: 50, Armor: 150,
		Speed: 15000, FuelConsumption: 300, StorageUsage: 4,
		RapidFire: map[string]int{"lightFighter": 6, "espionageProbe": 5, "missile": 10},
	},
	"battleship": {
		Cost: CostEntry{Metal: 35000, Crystal: 15000}, BuildTime: 210,
		CargoCapacity: 1500, Attack: 500, Shield: 200, Armor: 600,
		Speed: 10000, FuelConsumption: 500, StorageUsage: 8,
		RapidFire: map[string]int{"espionageProbe": 5},
	},
	"battlecruiser": {
		Cost: CostEntry{Metal: 30000, Crystal: 40000, Deuterium: 15000}, BuildTime: 270,
		CargoCapacity: 750, Attack: 700, Shield: 400, Armor: 700,
		Speed: 10000, FuelConsumption: 250, StorageUsage: 4,
		RapidFire: map[string]int{"espionageProbe": 5, "bomber": 2, "lightFighter": 10, "heavyFighter": 4},
	},
	"bomber": {
		Cost: CostEntry{Metal: 50000, Crystal: 25000, Deuterium: 15000}, BuildTime: 300,
		CargoCapacity: 500, Attack: 1000, Shield: 100, Armor: 750,
		Speed: 9000, FuelConsumption: 700, StorageUsage: 4,
		RapidFire: map[string]int{"espionageProbe": 5, "rocketLauncher": 10, "lightLaser": 5},
	},
	"destroyer": {
		Cost: CostEntry{Metal: 60000, Crystal: 50000, Deuterium: 15000}, BuildTime: 360,
		CargoCapacity: 2000, Attack: 2000, Shield: 500, Armor: 1100,
		Speed: 8000, FuelConsumption: 1000, StorageUsage: 8,
		RapidFire: map[string]int{"espionageProbe": 5, "bomber": 4},
	},
	"smallCargo": {
		Cost: CostEntry{Metal: 2000, Crystal: 2000}, BuildTime: 30,
		CargoCapacity: 5000, Attack: 5, Shield: 10, Armor: 10,
		Speed: 5000, FuelConsumption: 20, StorageUsage: 1,
		RapidFire: map[string]int{"espionageProbe": 5},
	},
	"largeCargo": {
		Cost: CostEntry{Metal: 6000, Crystal: 6000}, BuildTime: 60,
		CargoCapacity: 25000, Attack: 5, Shield: 10, Armor: 25,
		Speed: 7500, FuelConsumption: 50, StorageUsage: 2,
		RapidFire: map[string]int{"espionageProbe": 5},
	},
	"colonyShip": {
		Cost: CostEntry{Metal: 10000, Crystal: 20000, Deuterium: 10000}, BuildTime: 300,
		CargoCapacity: 7500, Attack: 5, Shield: 100, Armor: 200,
		Speed: 2500, FuelConsumption: 1000, StorageUsage: 10,
	},
	"recycler": {
		Cost: CostEntry{Metal: 10000, Crystal: 6000, Deuterium: 2000}, BuildTime: 120,
		CargoCapacity: 20000, Attack: 1, Shield: 10, Armor: 40,
		Speed: 4000, FuelConsumption: 40, StorageUsage: 2,
		RapidFire: map[string]int{"espionageProbe": 5},
	},
	"espionageProbe": {
		Cost: CostEntry{Crystal: 1000}, BuildTime: 5,
		CargoCapacity: 5, Attack: 1, Shield: 1, Armor: 1,
		Speed: 100000000, FuelConsumption: 1, StorageUsage: 1,
	},
	"solarSatellite": {
		Cost: CostEntry{Crystal: 2000, Deuterium: 500}, BuildTime: 10,
		CargoCapacity: 0, Attack: 0, Shield: 0, Armor: 0,
		Speed: 0, FuelConsumption: 0, StorageUsage: 1,
	},
	"darkMatterHarvester": {
		Cost: CostEntry{Metal: 5000, Crystal: 2500, Deuterium: 1000}, BuildTime: 60,
		CargoCapacity: 500, Attack: 0, Shield: 0, Armor: 0,
		Speed: 0, FuelConsumption: 0, StorageUsage: 1,
	},
	"deathstar": {
		Cost: CostEntry{Metal: 5000000, Crystal: 4000000, Deuterium: 1000000}, BuildTime: 600,
		CargoCapacity: 1000000, Attack: 200000, Shield: 50000, Armor: 800000,
		Speed: 100, FuelConsumption: 1, StorageUsage: 1,
		RapidFire: map[string]int{
			"lightFighter": 200, "heavyFighter": 125, "cruiser": 33,
			"battleship": 30, "battlecruiser": 250, "bomber": 25,
			"destroyer": 5, "smallCargo": 125, "largeCargo": 250,
			"colonyShip": 250, "recycler": 125, "espionageProbe": 1250,
			"solarSatellite": 1250, "darkMatterHarvester": 125,
			"rocketLauncher": 200, "lightLaser": 200, "heavyLaser": 100,
			"gaussCannon": 10, "ionCannon": 100, "plasmaTurret": 10,
			"smallShieldDome": 200, "largeShieldDome": 200,
		},
	},
}

// DefenseDefs contains all defense definitions.
var DefenseDefs = map[string]DefenseDef{
	"rocketLauncher": {
		Cost: CostEntry{Metal: 2000}, BuildTime: 10,
		Attack: 50, Shield: 20, Armor: 20,
	},
	"lightLaser": {
		Cost: CostEntry{Metal: 1500, Crystal: 500}, BuildTime: 10,
		Attack: 100, Shield: 25, Armor: 80,
	},
	"heavyLaser": {
		Cost: CostEntry{Metal: 6000, Crystal: 2000}, BuildTime: 30,
		Attack: 250, Shield: 100, Armor: 200,
	},
	"gaussCannon": {
		Cost: CostEntry{Metal: 20000, Crystal: 15000, Deuterium: 2000}, BuildTime: 60,
		Attack: 1100, Shield: 500, Armor: 800,
	},
	"ionCannon": {
		Cost: CostEntry{Metal: 5000, Crystal: 3000}, BuildTime: 30,
		Attack: 500, Shield: 500, Armor: 500,
	},
	"plasmaTurret": {
		Cost: CostEntry{Metal: 50000, Crystal: 50000, Deuterium: 30000}, BuildTime: 180,
		Attack: 3000, Shield: 300, Armor: 1000,
	},
	"smallShieldDome": {
		Cost: CostEntry{Metal: 10000, Crystal: 10000}, BuildTime: 120,
		Attack: 1, Shield: 2000, Armor: 2000,
	},
	"largeShieldDome": {
		Cost: CostEntry{Metal: 50000, Crystal: 50000}, BuildTime: 600,
		Attack: 1, Shield: 10000, Armor: 10000,
	},
	"antiBallisticMissile": {
		Cost: CostEntry{Metal: 8000}, BuildTime: 10,
		Attack: 1, Shield: 1, Armor: 1,
	},
	"interplanetaryMissile": {
		Cost: CostEntry{Metal: 12500, Crystal: 2500, Deuterium: 10000}, BuildTime: 30,
		Attack: 12000, Shield: 1, Armor: 1,
	},
	"planetaryShield": {
		Cost: CostEntry{Metal: 100000, Crystal: 50000, Deuterium: 50000}, BuildTime: 600,
		Attack: 1, Shield: 50000, Armor: 50000,
	},
}
