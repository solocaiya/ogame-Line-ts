package engine

// BuildingDefs contains all building definitions.
// Values match client src/config/gameConfig.ts exactly.
var BuildingDefs = map[string]BuildingDef{
	"metalMine":              {BaseCost: CostEntry{Metal: 60, Crystal: 15}, BaseTime: 15, CostMultiplier: 1.5, SpaceUsage: 1},
	"crystalMine":            {BaseCost: CostEntry{Metal: 48, Crystal: 24}, BaseTime: 15, CostMultiplier: 1.6, SpaceUsage: 1},
	"deuteriumSynthesizer":   {BaseCost: CostEntry{Metal: 225, Crystal: 75}, BaseTime: 20, CostMultiplier: 1.5, SpaceUsage: 2},
	"solarPlant":             {BaseCost: CostEntry{Metal: 75, Crystal: 30}, BaseTime: 15, CostMultiplier: 1.5, SpaceUsage: 2},
	"fusionReactor":          {BaseCost: CostEntry{Metal: 900, Crystal: 360, Deuterium: 180}, BaseTime: 30, CostMultiplier: 1.8, SpaceUsage: 4},
	"roboticsFactory":        {BaseCost: CostEntry{Metal: 400, Crystal: 120, Deuterium: 200}, BaseTime: 40, CostMultiplier: 2.0, SpaceUsage: 4},
	"naniteFactory":          {BaseCost: CostEntry{Metal: 1000000, Crystal: 500000, Deuterium: 100000}, BaseTime: 240, CostMultiplier: 2.0, SpaceUsage: 8},
	"shipyard":               {BaseCost: CostEntry{Metal: 400, Crystal: 200, Deuterium: 100}, BaseTime: 30, CostMultiplier: 2.0, SpaceUsage: 5},
	"hangar":                 {BaseCost: CostEntry{Metal: 200, Crystal: 100, Deuterium: 50}, BaseTime: 20, CostMultiplier: 1.8, SpaceUsage: 3},
	"researchLab":            {BaseCost: CostEntry{Metal: 200, Crystal: 400, Deuterium: 200}, BaseTime: 30, CostMultiplier: 2.0, SpaceUsage: 3},
	"metalStorage":           {BaseCost: CostEntry{Metal: 1000}, BaseTime: 15, CostMultiplier: 2.0, SpaceUsage: 1},
	"crystalStorage":         {BaseCost: CostEntry{Metal: 1000, Crystal: 500}, BaseTime: 15, CostMultiplier: 2.0, SpaceUsage: 1},
	"deuteriumTank":          {BaseCost: CostEntry{Metal: 1000, Crystal: 1000}, BaseTime: 15, CostMultiplier: 2.0, SpaceUsage: 1},
	"darkMatterCollector":    {BaseCost: CostEntry{Metal: 50000, Crystal: 100000, Deuterium: 50000}, BaseTime: 90, CostMultiplier: 2.0, SpaceUsage: 6},
	"darkMatterTank":         {BaseCost: CostEntry{Metal: 10000, Crystal: 10000, Deuterium: 5000}, BaseTime: 20, CostMultiplier: 2.0, SpaceUsage: 2},
	"missileSilo":            {BaseCost: CostEntry{Metal: 20000, Crystal: 20000, Deuterium: 1000}, BaseTime: 45, CostMultiplier: 2.0, SpaceUsage: 5},
	"terraformer":            {BaseCost: CostEntry{Crystal: 50000, Deuterium: 100000}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 5},
	"lunarBase":              {BaseCost: CostEntry{Metal: 8000, Crystal: 8000, Deuterium: 4000}, BaseTime: 45, CostMultiplier: 2.0, SpaceUsage: 0},
	"sensorPhalanx":          {BaseCost: CostEntry{Metal: 20000, Crystal: 40000, Deuterium: 20000}, BaseTime: 60, CostMultiplier: 2.0, SpaceUsage: 6},
	"jumpGate":               {BaseCost: CostEntry{Metal: 2000000, Crystal: 4000000, Deuterium: 2000000, DarkMatter: 50000}, BaseTime: 240, CostMultiplier: 2.0, SpaceUsage: 10},
	"planetDestroyerFactory": {BaseCost: CostEntry{Metal: 5000000, Crystal: 4000000, Deuterium: 1000000, DarkMatter: 100000}, BaseTime: 300, CostMultiplier: 2.0, SpaceUsage: 15},
	"geoResearchStation":     {BaseCost: CostEntry{Metal: 50000, Crystal: 30000, Deuterium: 20000}, BaseTime: 60, CostMultiplier: 1.8, SpaceUsage: 4},
	"deepDrillingFacility":   {BaseCost: CostEntry{Metal: 100000, Crystal: 80000, Deuterium: 50000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 6},
	"university":             {BaseCost: CostEntry{Metal: 200000, Crystal: 100000, Deuterium: 50000}, BaseTime: 120, CostMultiplier: 2.0, SpaceUsage: 8},
}

// ShipDefs contains all ship definitions with combat stats.
// Values match client src/config/gameConfig.ts exactly.
var ShipDefs = map[string]ShipDef{
	"lightFighter": {
		Cost: CostEntry{Metal: 3000, Crystal: 1000}, BuildTime: 20,
		CargoCapacity: 50, Attack: 50, Shield: 10, Armor: 400,
		Speed: 12500, FuelConsumption: 20, StorageUsage: 1,
		RapidFire: map[string]int{"espionageProbe": 5, "solarSatellite": 5},
	},
	"heavyFighter": {
		Cost: CostEntry{Metal: 6000, Crystal: 4000}, BuildTime: 30,
		CargoCapacity: 100, Attack: 150, Shield: 25, Armor: 1000,
		Speed: 10000, FuelConsumption: 75, StorageUsage: 2,
		RapidFire: map[string]int{"smallCargo": 3, "espionageProbe": 5, "solarSatellite": 5},
	},
	"cruiser": {
		Cost: CostEntry{Metal: 20000, Crystal: 7000, Deuterium: 2000}, BuildTime: 60,
		CargoCapacity: 800, Attack: 400, Shield: 50, Armor: 2700,
		Speed: 15000, FuelConsumption: 300, StorageUsage: 4,
		RapidFire: map[string]int{"lightFighter": 6, "espionageProbe": 5, "solarSatellite": 5, "rocketLauncher": 10},
	},
	"battleship": {
		Cost: CostEntry{Metal: 45000, Crystal: 15000}, BuildTime: 90,
		CargoCapacity: 1500, Attack: 1200, Shield: 300, Armor: 10000,
		Speed: 10000, FuelConsumption: 500, StorageUsage: 8,
		RapidFire: map[string]int{"espionageProbe": 5, "solarSatellite": 5},
	},
	"battlecruiser": {
		Cost: CostEntry{Metal: 30000, Crystal: 40000, Deuterium: 15000}, BuildTime: 70,
		CargoCapacity: 750, Attack: 700, Shield: 400, Armor: 7000,
		Speed: 10000, FuelConsumption: 250, StorageUsage: 20,
		RapidFire: map[string]int{"smallCargo": 3, "largeCargo": 3, "heavyFighter": 4, "cruiser": 4, "battleship": 7, "espionageProbe": 5, "solarSatellite": 5},
	},
	"bomber": {
		Cost: CostEntry{Metal: 50000, Crystal: 25000, Deuterium: 15000}, BuildTime: 100,
		CargoCapacity: 500, Attack: 1000, Shield: 500, Armor: 7500,
		Speed: 4000, FuelConsumption: 700, StorageUsage: 35,
		RapidFire: map[string]int{
			"espionageProbe": 5, "solarSatellite": 5,
			"rocketLauncher": 20, "lightLaser": 20, "heavyLaser": 10,
			"ionCannon": 10, "gaussCannon": 5, "plasmaTurret": 5,
		},
	},
	"destroyer": {
		Cost: CostEntry{Metal: 15000, Crystal: 5000, Deuterium: 1500}, BuildTime: 45,
		CargoCapacity: 500, Attack: 250, Shield: 40, Armor: 2000,
		Speed: 18000, FuelConsumption: 200, StorageUsage: 12,
		RapidFire: map[string]int{"lightFighter": 4, "espionageProbe": 10, "solarSatellite": 10, "smallCargo": 2, "rocketLauncher": 5},
	},
	"smallCargo": {
		Cost: CostEntry{Metal: 2000, Crystal: 2000}, BuildTime: 15,
		CargoCapacity: 5000, Attack: 5, Shield: 10, Armor: 400,
		Speed: 5000, FuelConsumption: 10, StorageUsage: 1,
	},
	"largeCargo": {
		Cost: CostEntry{Metal: 6000, Crystal: 6000}, BuildTime: 30,
		CargoCapacity: 25000, Attack: 5, Shield: 25, Armor: 1200,
		Speed: 7500, FuelConsumption: 50, StorageUsage: 2,
	},
	"colonyShip": {
		Cost: CostEntry{Metal: 10000, Crystal: 20000, Deuterium: 10000}, BuildTime: 120,
		CargoCapacity: 7500, Attack: 50, Shield: 100, Armor: 3000,
		Speed: 2500, FuelConsumption: 1000, StorageUsage: 10,
	},
	"recycler": {
		Cost: CostEntry{Metal: 10000, Crystal: 6000, Deuterium: 2000}, BuildTime: 60,
		CargoCapacity: 20000, Attack: 1, Shield: 10, Armor: 1600,
		Speed: 2000, FuelConsumption: 300, StorageUsage: 2,
	},
	"espionageProbe": {
		Cost: CostEntry{Crystal: 1000}, BuildTime: 5,
		CargoCapacity: 5, Attack: 0, Shield: 0, Armor: 100,
		Speed: 100000000, FuelConsumption: 1, StorageUsage: 2,
	},
	"solarSatellite": {
		Cost: CostEntry{Crystal: 2000, Deuterium: 500}, BuildTime: 10,
		CargoCapacity: 0, Attack: 1, Shield: 1, Armor: 200,
		Speed: 1, FuelConsumption: 0, StorageUsage: 1,
	},
	"darkMatterHarvester": {
		Cost: CostEntry{Metal: 100000, Crystal: 150000, Deuterium: 50000}, BuildTime: 120,
		CargoCapacity: 1000, Attack: 10, Shield: 50, Armor: 2000,
		Speed: 5000, FuelConsumption: 500, StorageUsage: 50,
	},
	"deathstar": {
		Cost: CostEntry{Metal: 50000000, Crystal: 40000000, Deuterium: 10000000, DarkMatter: 20000}, BuildTime: 600,
		CargoCapacity: 1000000, Attack: 200000, Shield: 50000, Armor: 900000,
		Speed: 100, FuelConsumption: 1, StorageUsage: 100,
		RapidFire: map[string]int{
			"smallCargo": 250, "largeCargo": 250, "lightFighter": 200,
			"heavyFighter": 100, "cruiser": 33, "battleship": 30,
			"battlecruiser": 15, "bomber": 25, "destroyer": 5,
			"colonyShip": 250, "recycler": 250, "espionageProbe": 1250,
			"solarSatellite": 1250, "darkMatterHarvester": 50,
			"rocketLauncher": 200, "lightLaser": 200, "heavyLaser": 100,
			"gaussCannon": 50, "ionCannon": 100, "plasmaTurret": 10,
			"smallShieldDome": 1250, "largeShieldDome": 1250,
		},
	},
}

// ResearchDefs contains all research technology definitions.
// Costs scale exponentially: floor(baseCost * costMultiplier^(level-1))
// Values match client src/config/gameConfig.ts exactly.
var ResearchDefs = map[string]ResearchDef{
	"energyTech":    {BaseCost: CostEntry{Crystal: 800, Deuterium: 400}, BaseTime: 30, CostMultiplier: 2.0},
	"laserTech":     {BaseCost: CostEntry{Metal: 200, Crystal: 100}, BaseTime: 60, CostMultiplier: 2.0},
	"ionTech":       {BaseCost: CostEntry{Metal: 1000, Crystal: 300, Deuterium: 100}, BaseTime: 60, CostMultiplier: 2.0},
	"hyperspaceTech": {BaseCost: CostEntry{Crystal: 4000, Deuterium: 2000}, BaseTime: 60, CostMultiplier: 2.0},
	"plasmaTechnology": {BaseCost: CostEntry{Metal: 2000, Crystal: 4000, Deuterium: 1000}, BaseTime: 60, CostMultiplier: 2.0},
	"computerTech":    {BaseCost: CostEntry{Crystal: 400, Deuterium: 600}, BaseTime: 60, CostMultiplier: 2.0},
	"espionageTech":   {BaseCost: CostEntry{Metal: 200, Crystal: 1000, Deuterium: 200}, BaseTime: 60, CostMultiplier: 2.0},
	"weaponsTech":     {BaseCost: CostEntry{Metal: 800, Crystal: 200}, BaseTime: 60, CostMultiplier: 2.0},
	"shieldingTech":   {BaseCost: CostEntry{Metal: 200, Crystal: 600}, BaseTime: 60, CostMultiplier: 2.0},
	"armorTech":       {BaseCost: CostEntry{Metal: 1000}, BaseTime: 60, CostMultiplier: 2.0},
	"astrophysics":    {BaseCost: CostEntry{Metal: 4000, Crystal: 8000, Deuterium: 4000}, BaseTime: 60, CostMultiplier: 1.75},
	"gravitonTech":    {BaseCost: CostEntry{DarkMatter: 100000}, BaseTime: 0, CostMultiplier: 3.0},
	"combustionDrive": {BaseCost: CostEntry{Metal: 400, Deuterium: 600}, BaseTime: 60, CostMultiplier: 2.0},
	"impulseDrive":    {BaseCost: CostEntry{Metal: 2000, Crystal: 4000, Deuterium: 600}, BaseTime: 60, CostMultiplier: 2.0},
	"hyperspaceDrive": {BaseCost: CostEntry{Metal: 10000, Crystal: 20000, Deuterium: 6000}, BaseTime: 60, CostMultiplier: 2.0},
	"darkMatterTechnology": {BaseCost: CostEntry{Metal: 100000, Crystal: 200000, Deuterium: 100000}, BaseTime: 180, CostMultiplier: 2.0},
	"terraformingTechnology": {BaseCost: CostEntry{Crystal: 20000, Deuterium: 40000}, BaseTime: 90, CostMultiplier: 2.0},
	"planetDestructionTech": {BaseCost: CostEntry{Metal: 4000000, Crystal: 8000000, Deuterium: 4000000, DarkMatter: 200000}, BaseTime: 300, CostMultiplier: 2.0},
	"miningTechnology": {BaseCost: CostEntry{Metal: 50000, Crystal: 30000, Deuterium: 20000}, BaseTime: 90, CostMultiplier: 1.8},
	"intergalacticResearchNetwork": {BaseCost: CostEntry{Metal: 240000, Crystal: 400000, Deuterium: 160000}, BaseTime: 180, CostMultiplier: 2.0},
	"mineralResearch":  {BaseCost: CostEntry{Metal: 60000, Crystal: 30000}, BaseTime: 60, CostMultiplier: 1.75},
	"crystalResearch":  {BaseCost: CostEntry{Metal: 40000, Crystal: 60000}, BaseTime: 60, CostMultiplier: 1.75},
	"fuelResearch":     {BaseCost: CostEntry{Crystal: 50000, Deuterium: 50000}, BaseTime: 60, CostMultiplier: 1.75},
}

// DefenseDefs contains all defense definitions.
// Values match client src/config/gameConfig.ts exactly.
var DefenseDefs = map[string]DefenseDef{
	"rocketLauncher": {
		Cost: CostEntry{Metal: 2000}, BuildTime: 10,
		Attack: 80, Shield: 20, Armor: 200,
	},
	"lightLaser": {
		Cost: CostEntry{Metal: 1500, Crystal: 500}, BuildTime: 12,
		Attack: 100, Shield: 25, Armor: 200,
	},
	"heavyLaser": {
		Cost: CostEntry{Metal: 6000, Crystal: 2000}, BuildTime: 20,
		Attack: 250, Shield: 100, Armor: 800,
	},
	"gaussCannon": {
		Cost: CostEntry{Metal: 20000, Crystal: 15000, Deuterium: 2000}, BuildTime: 35,
		Attack: 1100, Shield: 200, Armor: 3500,
	},
	"ionCannon": {
		Cost: CostEntry{Metal: 2000, Crystal: 6000}, BuildTime: 30,
		Attack: 150, Shield: 500, Armor: 800,
	},
	"plasmaTurret": {
		Cost: CostEntry{Metal: 50000, Crystal: 50000, Deuterium: 30000}, BuildTime: 60,
		Attack: 3000, Shield: 300, Armor: 10000,
	},
	"smallShieldDome": {
		Cost: CostEntry{Metal: 10000, Crystal: 10000}, BuildTime: 30,
		Attack: 1, Shield: 2000, Armor: 2000,
	},
	"largeShieldDome": {
		Cost: CostEntry{Metal: 50000, Crystal: 50000}, BuildTime: 60,
		Attack: 1, Shield: 10000, Armor: 10000,
	},
	"antiBallisticMissile": {
		Cost: CostEntry{Metal: 8000, Deuterium: 2000}, BuildTime: 20,
		Attack: 1, Shield: 1, Armor: 800,
	},
	"interplanetaryMissile": {
		Cost: CostEntry{Metal: 12500, Crystal: 2500, Deuterium: 10000}, BuildTime: 30,
		Attack: 12000, Shield: 1, Armor: 1500,
	},
	"planetaryShield": {
		Cost: CostEntry{Metal: 2000000, Crystal: 2000000, Deuterium: 1000000, DarkMatter: 50000}, BuildTime: 180,
		Attack: 1, Shield: 100000, Armor: 100000,
	},
}
