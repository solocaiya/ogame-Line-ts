package engine

import "time"

// VIP level bitmask values.
// vip_level is stored as a bitmask so both cards can be active simultaneously:
//
//	bit 0 (1) = small monthly card active
//	bit 1 (2) = large monthly card active
//
// When a card is purchased, OR its bit into vip_level. Each card has its own
// expiry column (subscription_expires_at for small, subscription2_expires_at
// for large). Expired cards are ignored at bonus-calculation time; the bit is
// cleared lazily when the player next logs in or claims daily DM.
const (
	VIPBitSmallMonthly = 1 // bit 0
	VIPBitLargeMonthly = 2 // bit 1
)

// VIPBonus holds all active VIP/subscription bonuses for a player.
// Zero values mean no bonus for that attribute.
type VIPBonus struct {
	BuildQueueBonus    int     // extra building queue slots
	ResearchQueueBonus int     // extra research queue slots
	BuildSpeedPct      float64 // percentage reduction in build time (e.g. 10 = 10% faster)
	ResearchSpeedPct   float64 // percentage reduction in research time
	AttackPct          float64 // attack power bonus %
	DefensePct         float64 // defense power bonus %
	FleetSpeedPct      float64 // fleet travel speed bonus %
	TradeBonusPct      float64 // trader exchange rate bonus % (e.g. 10 = 10% better rates)
	AutoUpgrade        bool    // auto-upgrade buildings when resources available
	MaxQueueCap        int     // extra max queue length (added to base cap)
}

// IsZero reports whether no VIP bonuses are active.
func (b VIPBonus) IsZero() bool {
	return b.BuildQueueBonus == 0 &&
		b.ResearchQueueBonus == 0 &&
		b.BuildSpeedPct == 0 &&
		b.ResearchSpeedPct == 0 &&
		b.AttackPct == 0 &&
		b.DefensePct == 0 &&
		b.FleetSpeedPct == 0 &&
		b.TradeBonusPct == 0 &&
		!b.AutoUpgrade &&
		b.MaxQueueCap == 0
}

// GetVIPBonus computes the active VIP bonuses for a player.
//
// vipLevel is a bitmask (see VIPBitSmallMonthly, VIPBitLargeMonthly).
// Each card's expiry is checked independently; an expired card contributes
// nothing even if its bit is still set.
func GetVIPBonus(vipLevel int, subExpiresAt, sub2ExpiresAt *time.Time) VIPBonus {
	now := time.Now()
	b := VIPBonus{}

	smallActive := vipLevel&VIPBitSmallMonthly != 0 && subExpiresAt != nil && now.Before(*subExpiresAt)
	largeActive := vipLevel&VIPBitLargeMonthly != 0 && sub2ExpiresAt != nil && now.Before(*sub2ExpiresAt)

	if smallActive {
		// 小月卡: +1 build queue, +10% build speed
		b.BuildQueueBonus += 1
		b.BuildSpeedPct += 10.0
		b.MaxQueueCap += 2
	}

	if largeActive {
		// 大月卡: +1 research queue, +10% research speed,
		// +5% attack, +5% defense, +25% fleet speed,
		// +10% trade bonus, auto-upgrade
		b.ResearchQueueBonus += 1
		b.ResearchSpeedPct += 10.0
		b.AttackPct += 5.0
		b.DefensePct += 5.0
		b.FleetSpeedPct += 25.0
		b.TradeBonusPct += 10.0
		b.AutoUpgrade = true
		b.MaxQueueCap += 3
	}

	return b
}

// HasActiveCard reports whether any VIP card is currently active.
func HasActiveCard(vipLevel int, subExpiresAt, sub2ExpiresAt *time.Time) bool {
	return !GetVIPBonus(vipLevel, subExpiresAt, sub2ExpiresAt).IsZero()
}

// ClearExpiredBits returns an updated vipLevel with expired card bits cleared.
// Use this when loading a player to keep the stored bitmask accurate.
func ClearExpiredBits(vipLevel int, subExpiresAt, sub2ExpiresAt *time.Time) int {
	now := time.Now()
	if subExpiresAt == nil || !now.Before(*subExpiresAt) {
		vipLevel &^= VIPBitSmallMonthly
	}
	if sub2ExpiresAt == nil || !now.Before(*sub2ExpiresAt) {
		vipLevel &^= VIPBitLargeMonthly
	}
	return vipLevel
}
