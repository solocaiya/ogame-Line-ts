package engine

import (
	"math"
	"math/rand"
)

// Combat effects configuration (matches client battleSimulator.ts)
const (
	ShieldBounceThreshold       = 0.01
	ShieldBouncePenetrationChance = 0.01
	ShieldRegenRate             = 0.7
	ArmorDamageRate             = 0.05
	MaxArmorDamage              = 0.3
	HeavyWeaponShieldPenetration = 0.15
	HeavyWeaponThreshold        = 5000
	MaxCombatRounds             = 6
	DefenseRepairRate           = 0.7
	DebrisFieldRatio            = 0.3 // 30% of lost ship/defense cost goes to debris
	PlunderRatio                = 0.5 // 50% of defender resources can be plundered
	MoonBaseChance              = 0.0
	MoonMaxChance               = 20.0
	MoonDebrisThreshold         = 100000  // 100k debris for moon chance
	MoonChancePerDebris         = 100000  // per 100k debris = +1%
)

// applyTechBonus applies weapon/shield/armor tech bonus.
func applyTechBonus(baseValue int, techLevel int) int {
	return int(math.Floor(float64(baseValue) * (1.0 + float64(techLevel)*0.1)))
}

// prepareCombatUnits converts ships and defenses to combat units.
func prepareCombatUnits(side BattleSide, isDefender bool) []CombatUnit {
	var units []CombatUnit

	for shipType, count := range side.Ships {
		if count <= 0 {
			continue
		}
		def, ok := ShipDefs[shipType]
		if !ok {
			continue
		}
		units = append(units, CombatUnit{
			Type:          shipType,
			Count:         count,
			Attack:        applyTechBonus(def.Attack, side.WeaponTech),
			Shield:        applyTechBonus(def.Shield, side.ShieldTech),
			Armor:         applyTechBonus(def.Armor, side.ArmorTech),
			RapidFire:     def.RapidFire,
			CurrentShield: applyTechBonus(def.Shield, side.ShieldTech),
			ArmorDamage:   0,
		})
	}

	if isDefender {
		for defType, count := range side.Defense {
			if count <= 0 {
				continue
			}
			def, ok := DefenseDefs[defType]
			if !ok {
				continue
			}
			units = append(units, CombatUnit{
				Type:          defType,
				Count:         count,
				Attack:        applyTechBonus(def.Attack, side.WeaponTech),
				Shield:        applyTechBonus(def.Shield, side.ShieldTech),
				Armor:         applyTechBonus(def.Armor, side.ArmorTech),
				CurrentShield: applyTechBonus(def.Shield, side.ShieldTech),
				ArmorDamage:   0,
			})
		}
	}

	return units
}

// calculateDamage computes damage from attacker to defender.
func calculateDamage(attacker, defender *CombatUnit) (destroyed int, damagedShield int, armorDamageDealt float64) {
	attackPower := attacker.Attack
	defenderCurrentShield := defender.CurrentShield
	armorDamageMultiplier := 1.0 - defender.ArmorDamage
	effectiveArmor := float64(defender.Armor) * armorDamageMultiplier

	// Heavy weapon penetration
	shieldPenetration := 0.0
	if attackPower >= HeavyWeaponThreshold {
		if rand.Float64() < HeavyWeaponShieldPenetration {
			shieldPenetration = 0.3 + rand.Float64()*0.2
		}
	}

	effectiveShield := float64(defenderCurrentShield) * (1.0 - shieldPenetration)

	// Shield bounce: if attack < 1% of shield, likely bounces
	if float64(attackPower) < effectiveShield*ShieldBounceThreshold {
		if rand.Float64() > ShieldBouncePenetrationChance {
			return 0, 0, 0
		}
	}

	remainingDamage := float64(attackPower)

	// Consume shield first
	if remainingDamage > effectiveShield {
		remainingDamage -= effectiveShield
		damagedShield = int(effectiveShield)
		defender.CurrentShield = 0
	} else {
		damagedShield = int(remainingDamage)
		defender.CurrentShield = int(float64(defenderCurrentShield) - remainingDamage)
		return 0, damagedShield, 0
	}

	// Armor damage accumulation
	if remainingDamage > 0 {
		armorDamageDealt = math.Min(ArmorDamageRate, MaxArmorDamage-defender.ArmorDamage)
		defender.ArmorDamage = math.Min(MaxArmorDamage, defender.ArmorDamage+armorDamageDealt)
	}

	// Check armor destruction
	if remainingDamage > effectiveArmor {
		destroyed = 1
	} else {
		destroyChance := remainingDamage / effectiveArmor
		if rand.Float64() < destroyChance {
			destroyed = 1
		}
	}

	return destroyed, damagedShield, armorDamageDealt
}

// executeAttack handles a single unit attacking, including rapid fire.
func executeAttack(attacker *CombatUnit, targets *[]CombatUnit, shipLosses, defenseLosses map[string]int) {
	if len(*targets) == 0 {
		return
	}

	targetIdx := rand.Intn(len(*targets))
	target := &(*targets)[targetIdx]

	destroyed, _, _ := calculateDamage(attacker, target)

	if destroyed > 0 {
		target.Count -= destroyed
		if _, isShip := ShipDefs[target.Type]; isShip {
			shipLosses[target.Type] += destroyed
		} else {
			defenseLosses[target.Type] += destroyed
		}
		if target.Count <= 0 {
			*targets = append((*targets)[:targetIdx], (*targets)[targetIdx+1:]...)
		}
	}

	// Rapid fire
	if attacker.RapidFire != nil && len(*targets) > 0 {
		if rf, ok := attacker.RapidFire[target.Type]; ok && rf > 1 {
			continueChance := float64(rf-1) / float64(rf)
			if rand.Float64() < continueChance {
				executeAttack(attacker, targets, shipLosses, defenseLosses)
			}
		}
	}
}

// executeRound runs one round of combat.
func executeRound(attackerUnits, defenderUnits *[]CombatUnit) (attackerLosses, defenderShipLosses, defenderDefenseLosses map[string]int) {
	attackerLosses = make(map[string]int)
	defenderShipLosses = make(map[string]int)
	defenderDefenseLosses = make(map[string]int)

	// Attackers fire at defenders (with rapid fire)
	for i := range *attackerUnits {
		a := &(*attackerUnits)[i]
		for j := 0; j < a.Count; j++ {
			if len(*defenderUnits) == 0 {
				break
			}
			executeAttack(a, defenderUnits, defenderShipLosses, defenderDefenseLosses)
		}
	}

	// Defenders fire at attackers (no rapid fire for defenses)
	for i := range *defenderUnits {
		d := &(*defenderUnits)[i]
		for j := 0; j < d.Count; j++ {
			if len(*attackerUnits) == 0 {
				break
			}
			targetIdx := rand.Intn(len(*attackerUnits))
			target := &(*attackerUnits)[targetIdx]

			destroyed, _, _ := calculateDamage(d, target)
			if destroyed > 0 {
				target.Count -= destroyed
				attackerLosses[target.Type] += destroyed
				if target.Count <= 0 {
					*attackerUnits = append((*attackerUnits)[:targetIdx], (*attackerUnits)[targetIdx+1:]...)
				}
			}
		}
	}

	return
}

// regenerateShields restores shields after a round.
func regenerateShields(units *[]CombatUnit) {
	for i := range *units {
		u := &(*units)[i]
		if u.CurrentShield < u.Shield {
			regenAmount := int(float64(u.Shield) * ShieldRegenRate)
			u.CurrentShield = min(u.Shield, u.CurrentShield+regenAmount)
		}
	}
}

// SimulateBattle runs a full combat simulation.
func SimulateBattle(attacker, defender BattleSide, maxRounds int) BattleResult {
	if maxRounds <= 0 {
		maxRounds = MaxCombatRounds
	}

	attackerUnits := prepareCombatUnits(attacker, false)
	defenderUnits := prepareCombatUnits(defender, true)

	totalAttackerLosses := make(map[string]int)
	totalDefenderShipLosses := make(map[string]int)
	totalDefenderDefenseLosses := make(map[string]int)

	rounds := 0
	for round := 0; round < maxRounds; round++ {
		if len(attackerUnits) == 0 || len(defenderUnits) == 0 {
			break
		}
		rounds++

		aLosses, dShipLosses, dDefLosses := executeRound(&attackerUnits, &defenderUnits)
		regenerateShields(&attackerUnits)
		regenerateShields(&defenderUnits)

		for k, v := range aLosses {
			totalAttackerLosses[k] += v
		}
		for k, v := range dShipLosses {
			totalDefenderShipLosses[k] += v
		}
		for k, v := range dDefLosses {
			totalDefenderDefenseLosses[k] += v
		}
	}

	// Repair 70% of lost defenses
	repairedDefense := make(map[string]int)
	for defType, lost := range totalDefenderDefenseLosses {
		repaired := int(math.Floor(float64(lost) * DefenseRepairRate))
		if repaired > 0 {
			repairedDefense[defType] = repaired
			totalDefenderDefenseLosses[defType] -= repaired
		}
	}

	// Build remaining maps
	attackerRemaining := make(map[string]int)
	for _, u := range attackerUnits {
		if u.Count > 0 {
			attackerRemaining[u.Type] = u.Count
		}
	}

	defenderFleetRemaining := make(map[string]int)
	defenderDefenseRemaining := make(map[string]int)
	for _, u := range defenderUnits {
		if u.Count > 0 {
			if _, isShip := ShipDefs[u.Type]; isShip {
				defenderFleetRemaining[u.Type] = u.Count
			} else {
				defenderDefenseRemaining[u.Type] = u.Count
			}
		}
	}
	// Add repaired defenses
	for defType, count := range repairedDefense {
		defenderDefenseRemaining[defType] += count
	}

	// Determine winner
	var winner string
	if len(attackerUnits) == 0 && len(defenderUnits) == 0 {
		winner = "draw"
	} else if len(attackerUnits) == 0 {
		winner = "defender"
	} else if len(defenderUnits) == 0 {
		winner = "attacker"
	} else {
		if maxRounds > MaxCombatRounds {
			// Battle-to-finish mode: compare armor
			attackerPower := 0
			for _, u := range attackerUnits {
				attackerPower += u.Count * u.Armor
			}
			defenderPower := 0
			for _, u := range defenderUnits {
				defenderPower += u.Count * u.Armor
			}
			if float64(attackerPower) > float64(defenderPower)*1.2 {
				winner = "attacker"
			} else if float64(defenderPower) > float64(attackerPower)*1.2 {
				winner = "defender"
			} else {
				winner = "draw"
			}
		} else {
			winner = "draw"
		}
	}

	// Calculate plunder (only when attacker wins)
	var plunder Resources
	if winner == "attacker" {
		plunder = CalculatePlunder(defender.DefenderResources, attackerRemaining)
	}

	// Calculate debris field
	debris := calculateDebrisField(totalAttackerLosses, totalDefenderShipLosses, totalDefenderDefenseLosses)

	// Moon chance
	totalDebris := debris.Metal + debris.Crystal
	moonChance := 0.0
	if totalDebris >= MoonDebrisThreshold {
		moonChance = math.Min(MoonMaxChance, MoonBaseChance+float64(totalDebris/MoonDebrisThreshold))
	}

	return BattleResult{
		Winner:                winner,
		Rounds:                rounds,
		AttackerLosses:        totalAttackerLosses,
		DefenderFleetLosses:   totalDefenderShipLosses,
		DefenderDefenseLosses: totalDefenderDefenseLosses,
		AttackerRemaining:     attackerRemaining,
		DefenderFleetRemaining:    defenderFleetRemaining,
		DefenderDefenseRemaining: defenderDefenseRemaining,
		Plunder:               plunder,
		DebrisField:           debris,
		MoonChance:            moonChance,
	}
}

// calculateDebrisField computes debris from losses.
func calculateDebrisField(attackerLosses, defenderShipLosses, defenderDefenseLosses map[string]int) Resources {
	var totalMetal, totalCrystal int64

	for shipType, count := range attackerLosses {
		if def, ok := ShipDefs[shipType]; ok {
			totalMetal += int64(float64(def.Cost.Metal) * float64(count) * DebrisFieldRatio)
			totalCrystal += int64(float64(def.Cost.Crystal) * float64(count) * DebrisFieldRatio)
		}
	}
	for shipType, count := range defenderShipLosses {
		if def, ok := ShipDefs[shipType]; ok {
			totalMetal += int64(float64(def.Cost.Metal) * float64(count) * DebrisFieldRatio)
			totalCrystal += int64(float64(def.Cost.Crystal) * float64(count) * DebrisFieldRatio)
		}
	}
	for defType, count := range defenderDefenseLosses {
		if def, ok := DefenseDefs[defType]; ok {
			totalMetal += int64(float64(def.Cost.Metal) * float64(count) * DebrisFieldRatio)
			totalCrystal += int64(float64(def.Cost.Crystal) * float64(count) * DebrisFieldRatio)
		}
	}

	return Resources{Metal: totalMetal, Crystal: totalCrystal}
}

// CalculatePlunder computes how much the attacker can plunder.
func CalculatePlunder(defenderResources Resources, attackerFleet map[string]int) Resources {
	totalCapacity := int64(0)
	for shipType, count := range attackerFleet {
		if def, ok := ShipDefs[shipType]; ok {
			totalCapacity += int64(def.CargoCapacity) * int64(count)
		}
	}

	available := Resources{
		Metal:     defenderResources.Metal / 2,
		Crystal:   defenderResources.Crystal / 2,
		Deuterium: defenderResources.Deuterium / 2,
	}

	totalAvailable := available.Metal + available.Crystal + available.Deuterium

	if totalCapacity >= totalAvailable {
		return available
	}

	if totalAvailable == 0 {
		return Resources{}
	}

	ratio := float64(totalCapacity) / float64(totalAvailable)
	return Resources{
		Metal:     int64(float64(available.Metal) * ratio),
		Crystal:   int64(float64(available.Crystal) * ratio),
		Deuterium: int64(float64(available.Deuterium) * ratio),
	}
}
