package engine

// Resources represents the 4 resources in the game.
type Resources struct {
	Metal      int64 `json:"metal"`
	Crystal    int64 `json:"crystal"`
	Deuterium  int64 `json:"deuterium"`
	DarkMatter int64 `json:"darkMatter"`
}

// Clone returns a deep copy.
func (r Resources) Clone() Resources {
	return Resources{
		Metal:      r.Metal,
		Crystal:    r.Crystal,
		Deuterium:  r.Deuterium,
		DarkMatter: r.DarkMatter,
	}
}

// Add returns r + other.
func (r Resources) Add(other Resources) Resources {
	return Resources{
		Metal:      r.Metal + other.Metal,
		Crystal:    r.Crystal + other.Crystal,
		Deuterium:  r.Deuterium + other.Deuterium,
		DarkMatter: r.DarkMatter + other.DarkMatter,
	}
}

// Sub returns r - other, clamped to 0.
func (r Resources) Sub(other Resources) Resources {
	return Resources{
		Metal:      max(0, r.Metal-other.Metal),
		Crystal:    max(0, r.Crystal-other.Crystal),
		Deuterium:  max(0, r.Deuterium-other.Deuterium),
		DarkMatter: max(0, r.DarkMatter-other.DarkMatter),
	}
}

// CanAfford returns true if r has at least as much as cost in every resource.
func (r Resources) CanAfford(cost Resources) bool {
	return r.Metal >= cost.Metal && r.Crystal >= cost.Crystal &&
		r.Deuterium >= cost.Deuterium && r.DarkMatter >= cost.DarkMatter
}

// Coordinate represents a position in the universe.
type Coordinate struct {
	Galaxy   int `json:"galaxy"`
	System   int `json:"system"`
	Position int `json:"position"`
}

// PlanetState holds the full state of a single planet.
type PlanetState struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Coordinate    Coordinate          `json:"coordinate"`
	Buildings     map[string]int      `json:"buildings"`
	Technologies  map[string]int      `json:"technologies"`
	Ships         map[string]int      `json:"ships"`
	Defenses      map[string]int      `json:"defenses"`
	Resources     Resources           `json:"resources"`
	StorageCap    Resources           `json:"storageCap"`
	Production    Resources           `json:"production"`    // per hour
	EnergyUsed    int64               `json:"energyUsed"`
	EnergyProd    int64               `json:"energyProd"`
	Efficiency    float64             `json:"efficiency"`    // 0.0-1.0
	BuildingQueue []BuildingQueueItem `json:"buildingQueue"`
	ResearchQueue []ResearchQueueItem `json:"researchQueue"`
	ShipQueue     []ShipQueueItem     `json:"shipQueue"`
	DefenseQueue  []DefenseQueueItem  `json:"defenseQueue"`
	LastUpdate    int64               `json:"lastUpdate"` // unix ms
}

// BuildingQueueItem represents a building construction in progress.
type BuildingQueueItem struct {
	Type        string `json:"type"`
	TargetLevel int    `json:"targetLevel"`
	EndTime     int64  `json:"endTime"` // unix ms
}

// ResearchQueueItem represents a research in progress.
type ResearchQueueItem struct {
	Type    string `json:"type"`
	EndTime int64  `json:"endTime"`
}

// ShipQueueItem represents ship production in progress.
type ShipQueueItem struct {
	Type    string `json:"type"`
	Count   int    `json:"count"`
	EndTime int64  `json:"endTime"`
}

// DefenseQueueItem represents defense production in progress.
type DefenseQueueItem struct {
	Type    string `json:"type"`
	Count   int    `json:"count"`
	EndTime int64  `json:"endTime"`
}

// FleetMission represents an active fleet mission.
type FleetMission struct {
	ID            string         `json:"id"`
	PlayerID      string         `json:"playerId"`
	MissionType   string         `json:"missionType"`
	Fleet         map[string]int `json:"fleet"`
	Cargo         Resources      `json:"cargo"`
	Origin        Coordinate     `json:"origin"`
	Target        Coordinate     `json:"target"`
	DepartureTime int64          `json:"departureTime"`
	ArrivalTime   int64          `json:"arrivalTime"`
	ReturnTime    int64          `json:"returnTime"`
	Status        string         `json:"status"` // "outbound" | "returning"
	Recalled      bool           `json:"recalled"`
	BattleToFinish bool          `json:"battleToFinish"` // 战斗到底模式：true=最多100回合，false=经典6回合
}

// PlayerSettings represents player-configurable game settings synced to the server.
type PlayerSettings struct {
	BattleToFinish bool `json:"battleToFinish"` // 战斗到底模式（100回合）
}

// PlayerState represents the full server-side state of a player.
type PlayerState struct {
	ID            string                  `json:"id"`
	UserID        string                  `json:"userId"`
	Planets       map[string]*PlanetState `json:"planets"`
	FleetMissions []FleetMission          `json:"fleetMissions"`
	GameSpeed     int                     `json:"gameSpeed"`
	DebrisField   Resources               `json:"debrisField"`
	Moons         map[string]string       `json:"moons"` // planetID -> moonID
	Settings      PlayerSettings          `json:"settings"`
}

// CombatUnit represents a unit in combat simulation.
type CombatUnit struct {
	Type          string
	Count         int
	Attack        int
	Shield        int
	Armor         int
	RapidFire     map[string]int
	CurrentShield int
	ArmorDamage   float64
}

// BattleSide represents one side in a battle.
type BattleSide struct {
	Ships            map[string]int
	Defense          map[string]int
	WeaponTech       int
	ShieldTech       int
	ArmorTech        int
	DefenderResources Resources // only used for defender side — plunder calculation
}

// BattleResult holds the outcome of a battle.
type BattleResult struct {
	Winner                string         `json:"winner"` // "attacker" | "defender" | "draw"
	Rounds                int            `json:"rounds"`
	AttackerLosses        map[string]int `json:"attackerLosses"`
	DefenderFleetLosses   map[string]int `json:"defenderFleetLosses"`
	DefenderDefenseLosses map[string]int `json:"defenderDefenseLosses"`
	AttackerRemaining     map[string]int `json:"attackerRemaining"`
	DefenderFleetRemaining    map[string]int `json:"defenderFleetRemaining"`
	DefenderDefenseRemaining map[string]int `json:"defenderDefenseRemaining"`
	Plunder               Resources      `json:"plunder"`
	DebrisField           Resources      `json:"debrisField"`
	MoonChance            float64        `json:"moonChance"`
}

// CostEntry represents the cost to build one unit.
type CostEntry struct {
	Metal     int64
	Crystal   int64
	Deuterium int64
}

// ShipDef holds ship statistics.
type ShipDef struct {
	Cost            CostEntry
	BuildTime       int // seconds base
	CargoCapacity   int
	Attack          int
	Shield          int
	Armor           int
	Speed           int
	FuelConsumption int
	StorageUsage    int
	RapidFire       map[string]int
}

// DefenseDef holds defense statistics.
type DefenseDef struct {
	Cost      CostEntry
	BuildTime int
	Attack    int
	Shield    int
	Armor     int
}

// BuildingDef holds building statistics.
type BuildingDef struct {
	BaseCost       CostEntry
	BaseTime       int // seconds
	CostMultiplier float64
	SpaceUsage     int
}

// ResearchDef holds research technology statistics.
type ResearchDef struct {
	BaseCost       CostEntry
	BaseTime       int // seconds
	CostMultiplier float64
}
