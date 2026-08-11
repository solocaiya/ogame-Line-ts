package engine

// RechargeProduct defines a dark matter purchase tier.
type RechargeProduct struct {
	ID         string `json:"id"`
	AmountRMB  int64  `json:"amountRMB"`
	DarkMatter int64  `json:"darkMatter"`
	Bonus      int64  `json:"bonus"` // bonus DM on top of base
	FirstBonus float64 `json:"firstBonus"` // first-recharge bonus multiplier (e.g. 0.5 = +50%)
}

// MonthlyCard defines a subscription card.
type MonthlyCard struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	AmountRMB    int64    `json:"amountRMB"`
	DailyDM      int64    `json:"dailyDM"`
	DurationDays int      `json:"durationDays"`
	Perks        []string `json:"perks"`
	VIPLevel     int      `json:"vipLevel"`
}

// GiftPack defines a time-limited gift pack.
type GiftPack struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	CostDM        int64           `json:"costDM"`
	OnceOnly      bool            `json:"onceOnly"`
	CooldownHours int             `json:"cooldownHours,omitempty"`
	Contents      GiftPackContents `json:"contents"`
}

// GiftPackContents lists what a gift pack gives.
type GiftPackContents struct {
	DarkMatter int64 `json:"darkMatter,omitempty"`
	Metal      int64 `json:"metal,omitempty"`
	Crystal    int64 `json:"crystal,omitempty"`
	Deuterium  int64 `json:"deuterium,omitempty"`
}

// GrowthFundStage defines a claimable stage in the growth fund.
type GrowthFundStage struct {
	ID          string `json:"id"`
	PointsReq   int64  `json:"pointsReq"` // total player points required
	RewardDM    int64  `json:"rewardDM"`
	Description string `json:"description"`
}

// RechargeProducts available for purchase.
var RechargeProducts = map[string]RechargeProduct{
	"dm_60":   {ID: "dm_60",   AmountRMB: 6,   DarkMatter: 60,   Bonus: 0,   FirstBonus: 0.5},
	"dm_300":  {ID: "dm_300",  AmountRMB: 30,  DarkMatter: 300,  Bonus: 0,   FirstBonus: 0.5},
	"dm_680":  {ID: "dm_680",  AmountRMB: 68,  DarkMatter: 680,  Bonus: 0,   FirstBonus: 0.5},
	"dm_1280": {ID: "dm_1280", AmountRMB: 128, DarkMatter: 1280, Bonus: 128, FirstBonus: 0.5},
	"dm_3280": {ID: "dm_3280", AmountRMB: 328, DarkMatter: 3280, Bonus: 492, FirstBonus: 0.5},
	"dm_6480": {ID: "dm_6480", AmountRMB: 648, DarkMatter: 6480, Bonus: 1620, FirstBonus: 0.5},
}

// RechargeProductList is the ordered list for display.
var RechargeProductList = []string{"dm_60", "dm_300", "dm_680", "dm_1280", "dm_3280", "dm_6480"}

// MonthlyCards available for purchase.
var MonthlyCards = map[string]MonthlyCard{
	"small_monthly": {
		ID: "small_monthly", Name: "小月卡", AmountRMB: 30,
		DailyDM: 100, DurationDays: 30, VIPLevel: 1,
		Perks: []string{"build_queue_plus_1", "build_speed_10pct"},
	},
	"large_monthly": {
		ID: "large_monthly", Name: "大月卡", AmountRMB: 68,
		DailyDM: 200, DurationDays: 30, VIPLevel: 2,
		Perks: []string{"attack_10pct", "defense_10pct", "fleet_speed_10pct"},
	},
}

// MonthlyCardList is the ordered list for display.
var MonthlyCardList = []string{"small_monthly", "large_monthly"}

// GrowthFundStages defines the 10 claimable stages.
// Total return: 22000 DM for ¥98.
var GrowthFundStages = []GrowthFundStage{
	{ID: "stage_1", PointsReq: 1000, RewardDM: 500, Description: "总积分达到 1,000"},
	{ID: "stage_2", PointsReq: 5000, RewardDM: 1000, Description: "总积分达到 5,000"},
	{ID: "stage_3", PointsReq: 10000, RewardDM: 1500, Description: "总积分达到 10,000"},
	{ID: "stage_4", PointsReq: 25000, RewardDM: 2000, Description: "总积分达到 25,000"},
	{ID: "stage_5", PointsReq: 50000, RewardDM: 2500, Description: "总积分达到 50,000"},
	{ID: "stage_6", PointsReq: 100000, RewardDM: 3000, Description: "总积分达到 100,000"},
	{ID: "stage_7", PointsReq: 200000, RewardDM: 3000, Description: "总积分达到 200,000"},
	{ID: "stage_8", PointsReq: 500000, RewardDM: 3000, Description: "总积分达到 500,000"},
	{ID: "stage_9", PointsReq: 1000000, RewardDM: 2500, Description: "总积分达到 1,000,000"},
	{ID: "stage_10", PointsReq: 2000000, RewardDM: 3000, Description: "总积分达到 2,000,000"},
}

// GrowthFundCostRMB is the one-time purchase price.
const GrowthFundCostRMB = 98

// GiftPacks available for purchase.
var GiftPacks = []GiftPack{
	{
		ID: "starter_pack", Name: "新手礼包", CostDM: 100, OnceOnly: true,
		Contents: GiftPackContents{DarkMatter: 500, Metal: 50000, Crystal: 30000},
	},
	{
		ID: "war_pack", Name: "战争礼包", CostDM: 300, OnceOnly: false, CooldownHours: 168,
		Contents: GiftPackContents{DarkMatter: 1500},
	},
	{
		ID: "resource_pack", Name: "资源礼包", CostDM: 200, OnceOnly: false, CooldownHours: 72,
		Contents: GiftPackContents{DarkMatter: 800, Metal: 200000, Crystal: 120000, Deuterium: 60000},
	},
}

// AccelerateCostPerHour is the DM cost to accelerate 60 minutes.
const AccelerateCostPerHour int64 = 1
