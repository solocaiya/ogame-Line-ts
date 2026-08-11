package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"ogame-server/engine"
	"ogame-server/gamestate"
	"ogame-server/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WalletHandler handles wallet, recharge, monthly card, and growth fund endpoints.
type WalletHandler struct {
	gameState *gamestate.GameState
	db        *sql.DB
	wsHub     *ws.Hub
}

// NewWalletHandler creates a new WalletHandler.
func NewWalletHandler(gameState *gamestate.GameState, db *sql.DB, wsHub *ws.Hub) *WalletHandler {
	return &WalletHandler{gameState: gameState, db: db, wsHub: wsHub}
}

// ─── Balance ─────────────────────────────────────────────────────────────────

// GetBalance returns the player's dark matter balance and related info.
func (h *WalletHandler) GetBalance(c *gin.Context) {
	playerID := c.GetString("user_id")

	var darkMatterBalance, consumptionPoints int64
	var vipLevel int
	var subscriptionExpiresAt, subscription2ExpiresAt sql.NullTime

	err := h.db.QueryRow(`
		SELECT dark_matter_balance, consumption_points, vip_level,
		       subscription_expires_at, subscription2_expires_at
		FROM users WHERE id = ?
	`, playerID).Scan(&darkMatterBalance, &consumptionPoints, &vipLevel, &subscriptionExpiresAt, &subscription2ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query balance"})
		return
	}

	resp := gin.H{
		"darkMatterBalance":  darkMatterBalance,
		"consumptionPoints":  consumptionPoints,
		"vipLevel":           vipLevel,
	}

	if subscriptionExpiresAt.Valid {
		resp["subscriptionExpiresAt"] = subscriptionExpiresAt.Time.Format(time.RFC3339)
	} else {
		resp["subscriptionExpiresAt"] = nil
	}
	if subscription2ExpiresAt.Valid {
		resp["subscription2ExpiresAt"] = subscription2ExpiresAt.Time.Format(time.RFC3339)
	} else {
		resp["subscription2ExpiresAt"] = nil
	}

	// Check monthly card daily claim availability
	resp["dailyDMReady"] = h.checkDailyDMReady(playerID)

	c.JSON(http.StatusOK, resp)
}

// checkDailyDMReady returns true if the player can claim daily DM from a monthly card.
// A player can claim if at least one card is active (not expired).
func (h *WalletHandler) checkDailyDMReady(playerID string) bool {
	var vipLevel int
	var expiresAt, expiresAt2 sql.NullTime
	var lastClaimAt sql.NullTime
	err := h.db.QueryRow(`
		SELECT vip_level, subscription_expires_at, subscription2_expires_at,
		       (SELECT MAX(created_at) FROM dark_matter_transactions WHERE user_id = ? AND type = 'daily_claim')
		FROM users WHERE id = ?
	`, playerID, playerID).Scan(&vipLevel, &expiresAt, &expiresAt2, &lastClaimAt)
	if err != nil {
		return false
	}
	now := time.Now()
	smallActive := vipLevel&engine.VIPBitSmallMonthly != 0 && expiresAt.Valid && now.Before(expiresAt.Time)
	largeActive := vipLevel&engine.VIPBitLargeMonthly != 0 && expiresAt2.Valid && now.Before(expiresAt2.Time)
	if !smallActive && !largeActive {
		return false
	}
	if !lastClaimAt.Valid {
		return true // never claimed
	}
	// Can claim if 22+ hours since last claim
	return now.Sub(lastClaimAt.Time) >= 22*time.Hour
}

// ─── Products ────────────────────────────────────────────────────────────────

// GetProducts returns available recharge products.
func (h *WalletHandler) GetProducts(c *gin.Context) {
	playerID := c.GetString("user_id")

	// Check if this is the player's first recharge (for first-bonus display)
	firstRecharge := false
	var count int
	err := h.db.QueryRow(`SELECT COUNT(*) FROM recharge_orders WHERE user_id = ? AND status = 'delivered'`, playerID).Scan(&count)
	if err == nil && count == 0 {
		firstRecharge = true
	}

	products := make([]engine.RechargeProduct, 0, len(engine.RechargeProductList))
	for _, id := range engine.RechargeProductList {
		p := engine.RechargeProducts[id]
		if !firstRecharge {
			p.FirstBonus = 0 // hide first bonus if already used
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"products":      products,
		"firstRecharge": firstRecharge,
	})
}

// ─── Recharge ────────────────────────────────────────────────────────────────

type createRechargeRequest struct {
	ProductID string `json:"productId" binding:"required"`
}

// CreateRecharge creates a new recharge order (mock payment flow).
func (h *WalletHandler) CreateRecharge(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req createRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	product, ok := engine.RechargeProducts[req.ProductID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product"})
		return
	}

	orderID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := h.db.Exec(`
		INSERT INTO recharge_orders (id, user_id, product_id, amount_rmb, dark_matter, status, created_at)
		VALUES (?, ?, ?, ?, ?, 'pending', ?)
	`, orderID, playerID, req.ProductID, product.AmountRMB, product.DarkMatter+product.Bonus, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orderId":   orderID,
		"amountRMB": product.AmountRMB,
		"darkMatter": product.DarkMatter + product.Bonus,
		"message":   "订单已创建，当前为测试模式，请调用 confirm-payment 完成支付",
	})
}

type confirmPaymentRequest struct {
	OrderID string `json:"orderId" binding:"required"`
}

// ConfirmPayment simulates payment confirmation (mock flow).
// In production, this would be triggered by a payment callback.
func (h *WalletHandler) ConfirmPayment(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req confirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Fetch the order
	var orderID, orderProductID, orderStatus string
	var orderDM int64
	err := h.db.QueryRow(`
		SELECT id, product_id, status, dark_matter FROM recharge_orders WHERE id = ? AND user_id = ?
	`, req.OrderID, playerID).Scan(&orderID, &orderProductID, &orderStatus, &orderDM)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query order"})
		return
	}
	if orderStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("order already %s", orderStatus)})
		return
	}

	// Check first-recharge bonus
	isFirstRecharge := false
	var prevCount int
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM recharge_orders WHERE user_id = ? AND status = 'delivered' AND id != ?`, playerID, orderID).Scan(&prevCount)
	if prevCount == 0 {
		isFirstRecharge = true
	}

	// Calculate bonus DM
	bonusDM := int64(0)
	if isFirstRecharge {
		if product, ok := engine.RechargeProducts[orderProductID]; ok {
			bonusDM = int64(float64(product.DarkMatter) * product.FirstBonus)
		}
	}

	totalDM := orderDM + bonusDM
	now := time.Now().UTC().Format(time.RFC3339)
	txID := uuid.New().String()

	// Transaction: update order, add balance, log transaction — all in a DB transaction
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// 1. Update order status
	_, err = tx.Exec(`UPDATE recharge_orders SET status = 'delivered', paid_at = ?, delivered_at = ? WHERE id = ?`, now, now, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
		return
	}

	// 2. Add dark matter balance
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance + ? WHERE id = ?`, totalDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update balance"})
		return
	}

	// 3. Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// 4. Log transaction
	txType := "recharge"
	if isFirstRecharge {
		txType = "recharge_first"
	}
	desc := fmt.Sprintf("充值 %d DM (%s)", product.DarkMatter+product.Bonus, product.ID)
	if isFirstRecharge {
		desc = "首充奖励 " + desc
	}
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, txID, playerID, totalDM, newBalance, txType, orderID, desc, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	// Notify via WebSocket
	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance":          newBalance,
			"consumptionPoints": 0, // unchanged
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"darkMatterAdded":  totalDM,
		"firstRechargeBonus": bonusDM,
		"newBalance":       newBalance,
	})
}

// ─── Transactions ────────────────────────────────────────────────────────────

// GetTransactions returns dark matter transaction history.
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	playerID := c.GetString("user_id")

	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := parseIntParam(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := parseIntParam(v); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, err := h.db.Query(`
		SELECT id, amount, balance_after, type, ref_id, created_at
		FROM dark_matter_transactions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, playerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query transactions"})
		return
	}
	defer rows.Close()

	transactions := make([]gin.H, 0)
	for rows.Next() {
		var id, txType, refID, createdAt string
		var amount, balanceAfter int64
		if err := rows.Scan(&id, &amount, &balanceAfter, &txType, &refID, &createdAt); err != nil {
			continue
		}
		transactions = append(transactions, gin.H{
			"id":           id,
			"amount":       amount,
			"balanceAfter": balanceAfter,
			"type":         txType,
			"refId":        refID,
			"createdAt":    createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// ─── Monthly Cards ───────────────────────────────────────────────────────────

// GetMonthlyCards returns available monthly cards and player's current status.
func (h *WalletHandler) GetMonthlyCards(c *gin.Context) {
	playerID := c.GetString("user_id")

	var vipLevel int
	var expiresAt, expiresAt2 sql.NullTime
	_ = h.db.QueryRow(`SELECT vip_level, subscription_expires_at, subscription2_expires_at FROM users WHERE id = ?`, playerID).
		Scan(&vipLevel, &expiresAt, &expiresAt2)

	now := time.Now()

	cards := make([]gin.H, 0, len(engine.MonthlyCardList))
	for _, id := range engine.MonthlyCardList {
		card := engine.MonthlyCards[id]

		// Determine active status and expiry per card
		active := false
		var cardExpiresAt *time.Time
		switch card.ID {
		case "small_monthly":
			active = vipLevel&engine.VIPBitSmallMonthly != 0 && expiresAt.Valid && now.Before(expiresAt.Time)
			if expiresAt.Valid {
				cardExpiresAt = &expiresAt.Time
			}
		case "large_monthly":
			active = vipLevel&engine.VIPBitLargeMonthly != 0 && expiresAt2.Valid && now.Before(expiresAt2.Time)
			if expiresAt2.Valid {
				cardExpiresAt = &expiresAt2.Time
			}
		}

		cardInfo := gin.H{
			"id":           card.ID,
			"name":         card.Name,
			"amountRMB":    card.AmountRMB,
			"dailyDM":      card.DailyDM,
			"durationDays": card.DurationDays,
			"perks":        card.Perks,
			"vipLevel":     card.VIPLevel,
			"active":       active,
		}
		if cardExpiresAt != nil {
			cardInfo["expiresAt"] = cardExpiresAt.Format(time.RFC3339)
		}
		cards = append(cards, cardInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"cards":      cards,
		"dailyReady": h.checkDailyDMReady(playerID),
	})
}

type buyMonthlyCardRequest struct {
	CardID string `json:"cardId" binding:"required"`
}

// BuyMonthlyCard purchases a monthly card using dark matter balance.
// In the design doc, cards are ¥30/¥68 — but for the mock flow, we charge DM equivalent.
// Small monthly = 300 DM, Large monthly = 680 DM (matching the recharge tiers).
func (h *WalletHandler) BuyMonthlyCard(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req buyMonthlyCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	card, ok := engine.MonthlyCards[req.CardID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown monthly card"})
		return
	}

	// Cost in DM (mock: 1 RMB = 100 DM)
	costDM := card.AmountRMB * 100

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Check balance
	var balance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query balance"})
		return
	}
	if balance < costDM {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter"})
		return
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, card.DurationDays)
	txID := uuid.New().String()

	// Determine which bit and expiry column to set based on card ID
	var bit int
	var expiryCol string
	switch card.ID {
	case "small_monthly":
		bit = engine.VIPBitSmallMonthly
		expiryCol = "subscription_expires_at"
	case "large_monthly":
		bit = engine.VIPBitLargeMonthly
		expiryCol = "subscription2_expires_at"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown card type"})
		return
	}

	// Deduct DM, OR the VIP bit, and set the card's expiry column
	updateSQL := fmt.Sprintf(
		`UPDATE users SET dark_matter_balance = dark_matter_balance - ?, vip_level = vip_level | ?, %s = ? WHERE id = ?`,
		expiryCol,
	)
	_, err = tx.Exec(updateSQL, costDM, bit, expiresAt, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// Add consumption points (1 per DM spent)
	_, err = tx.Exec(`UPDATE users SET consumption_points = consumption_points + ? WHERE id = ?`, costDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update consumption points"})
		return
	}

	// Log transaction
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, 'monthly_card', ?, ?, ?)
	`, txID, playerID, -costDM, newBalance, card.ID, fmt.Sprintf("购买%s (¥%d)", card.Name, card.AmountRMB), now.Format(time.RFC3339))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	// Read back updated vip_level and both expiry columns for the WS event
	var newVIPLevel int
	var newSubExpires, newSub2Expires sql.NullTime
	_ = h.db.QueryRow(`SELECT vip_level, subscription_expires_at, subscription2_expires_at FROM users WHERE id = ?`, playerID).
		Scan(&newVIPLevel, &newSubExpires, &newSub2Expires)

	// Sync VIP state to in-memory PlayerState so engine calculations pick it up immediately.
	var sub1, sub2 *time.Time
	if newSubExpires.Valid {
		t := newSubExpires.Time
		sub1 = &t
	}
	if newSub2Expires.Valid {
		t := newSub2Expires.Time
		sub2 = &t
	}
	_ = h.gameState.UpdatePlayer(playerID, func(player *engine.PlayerState) error {
		player.VIPLevel = newVIPLevel
		player.SubExpiresAt = sub1
		player.Sub2ExpiresAt = sub2
		return nil
	})

	// Notify
	if h.wsHub != nil {
		resp := gin.H{
			"balance":  newBalance,
			"vipLevel": newVIPLevel,
		}
		if newSubExpires.Valid {
			resp["subscriptionExpiresAt"] = newSubExpires.Time.Format(time.RFC3339)
		}
		if newSub2Expires.Valid {
			resp["subscription2ExpiresAt"] = newSub2Expires.Time.Format(time.RFC3339)
		}
		h.wsHub.SendTo(playerID, "dmBalanceChanged", resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"cardId":     card.ID,
		"costDM":     costDM,
		"expiresAt":  expiresAt.Format(time.RFC3339),
		"newBalance": newBalance,
	})
}

// ClaimDailyDM claims the daily DM from an active monthly card.
func (h *WalletHandler) ClaimDailyDM(c *gin.Context) {
	playerID := c.GetString("user_id")

	// Check that at least one card is active
	var vipLevel int
	var expiresAt, expiresAt2 sql.NullTime
	err := h.db.QueryRow(`SELECT vip_level, subscription_expires_at, subscription2_expires_at FROM users WHERE id = ?`, playerID).
		Scan(&vipLevel, &expiresAt, &expiresAt2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to query subscription status"})
		return
	}

	now := time.Now()
	smallActive := vipLevel&engine.VIPBitSmallMonthly != 0 && expiresAt.Valid && now.Before(expiresAt.Time)
	largeActive := vipLevel&engine.VIPBitLargeMonthly != 0 && expiresAt2.Valid && now.Before(expiresAt2.Time)
	if !smallActive && !largeActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active monthly card"})
		return
	}

	// Check cooldown (22h)
	var lastClaim sql.NullTime
	_ = h.db.QueryRow(`
		SELECT MAX(created_at) FROM dark_matter_transactions WHERE user_id = ? AND type = 'daily_claim'
	`, playerID).Scan(&lastClaim)
	if lastClaim.Valid && time.Since(lastClaim.Time) < 22*time.Hour {
		remaining := 22*time.Hour - time.Since(lastClaim.Time)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "daily DM already claimed, please try again later",
			"retryAfter": fmt.Sprintf("%.0f minutes", remaining.Minutes()),
		})
		return
	}

	// Sum daily DM from all active cards (both can stack)
	dailyDM := int64(0)
	if smallActive {
		dailyDM += engine.MonthlyCards["small_monthly"].DailyDM
	}
	if largeActive {
		dailyDM += engine.MonthlyCards["large_monthly"].DailyDM
	}

	now := time.Now().UTC().Format(time.RFC3339)
	txID := uuid.New().String()

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Add DM
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance + ? WHERE id = ?`, dailyDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update balance"})
		return
	}

	// Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// Log transaction
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, 'daily_claim', ?, ?, ?)
	`, txID, playerID, dailyDM, newBalance, fmt.Sprintf("vip_%d", vipLevel), fmt.Sprintf("每日领取 %d DM", dailyDM), now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{
			"balance": newBalance,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"dmClaimed":  dailyDM,
		"newBalance": newBalance,
	})
}

// ─── Growth Fund ─────────────────────────────────────────────────────────────

// GetGrowthFund returns the growth fund status and claimable stages.
func (h *WalletHandler) GetGrowthFund(c *gin.Context) {
	playerID := c.GetString("user_id")

	// Check if purchased
	var purchasedAt sql.NullString
	var totalClaimed int64
	_ = h.db.QueryRow(`SELECT purchased_at, total_claimed FROM growth_fund WHERE user_id = ?`, playerID).
		Scan(&purchasedAt, &totalClaimed)

	// Get player total points from leaderboard
	var totalPoints int64
	_ = h.db.QueryRow(`SELECT total_points FROM leaderboard WHERE user_id = ?`, playerID).Scan(&totalPoints)

	// Build stage list with claim status
	stages := make([]gin.H, 0, len(engine.GrowthFundStages))
	for _, stage := range engine.GrowthFundStages {
		claimed := totalClaimed >= stage.RewardDM // simplified check
		achievable := totalPoints >= stage.PointsReq

		// More precise: check if this specific stage was claimed
		// We track total_claimed, so we check by summing rewards up to this stage
		stageClaimed := false
		if purchasedAt.Valid {
			// Calculate cumulative reward up to this stage
			cumulative := int64(0)
			for _, s := range engine.GrowthFundStages {
				if s.ID == stage.ID {
					cumulative += s.RewardDM
					break
				}
				cumulative += s.RewardDM
			}
			stageClaimed = totalClaimed >= cumulative
		}

		stages = append(stages, gin.H{
			"id":          stage.ID,
			"pointsReq":   stage.PointsReq,
			"rewardDM":    stage.RewardDM,
			"description": stage.Description,
			"claimed":     stageClaimed,
			"achievable":  achievable,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"purchased":    purchasedAt.Valid,
		"purchasedAt":  purchasedAt.String,
		"totalClaimed": totalClaimed,
		"totalPoints":  totalPoints,
		"stages":       stages,
		"costRMB":      engine.GrowthFundCostRMB,
	})
}

// BuyGrowthFund purchases the growth fund (mock: costs DM equivalent of ¥98).
func (h *WalletHandler) BuyGrowthFund(c *gin.Context) {
	playerID := c.GetString("user_id")

	// Check not already purchased
	var exists int
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM growth_fund WHERE user_id = ?`, playerID).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "growth fund already purchased"})
		return
	}

	// Cost in DM (mock: 1 RMB = 100 DM)
	costDM := int64(engine.GrowthFundCostRMB * 100)

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Check balance
	var balance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query balance"})
		return
	}
	if balance < costDM {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	txID := uuid.New().String()

	// Deduct DM
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance - ? WHERE id = ?`, costDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct DM"})
		return
	}

	// Add consumption points
	_, err = tx.Exec(`UPDATE users SET consumption_points = consumption_points + ? WHERE id = ?`, costDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update points"})
		return
	}

	// Create growth fund record
	_, err = tx.Exec(`INSERT INTO growth_fund (user_id, purchased_at, total_claimed) VALUES (?, ?, 0)`, playerID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create growth fund"})
		return
	}

	// Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// Log transaction
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, 'growth_fund', 'purchase', ?, ?)
	`, txID, playerID, -costDM, newBalance, "购买成长基金", now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{"balance": newBalance})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"costDM":     costDM,
		"newBalance": newBalance,
	})
}

// ClaimGrowthFund claims a specific stage reward from the growth fund.
func (h *WalletHandler) ClaimGrowthFund(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		StageID string `json:"stageId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Find the stage
	var stage *engine.GrowthFundStage
	for i := range engine.GrowthFundStages {
		if engine.GrowthFundStages[i].ID == req.StageID {
			stage = &engine.GrowthFundStages[i]
			break
		}
	}
	if stage == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown stage"})
		return
	}

	// Check growth fund was purchased
	var totalClaimed int64
	err := h.db.QueryRow(`SELECT total_claimed FROM growth_fund WHERE user_id = ?`, playerID).Scan(&totalClaimed)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "growth fund not purchased"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query growth fund"})
		return
	}

	// Check this stage hasn't been claimed yet
	// Calculate cumulative reward up to and including this stage
	cumulative := int64(0)
	for _, s := range engine.GrowthFundStages {
		cumulative += s.RewardDM
		if s.ID == stage.ID {
			break
		}
	}
	if totalClaimed >= cumulative {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage already claimed"})
		return
	}

	// Check player points meet requirement
	var totalPoints int64
	_ = h.db.QueryRow(`SELECT total_points FROM leaderboard WHERE user_id = ?`, playerID).Scan(&totalPoints)
	if totalPoints < stage.PointsReq {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient points"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	txID := uuid.New().String()

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Add DM reward
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance + ? WHERE id = ?`, stage.RewardDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add DM"})
		return
	}

	// Update growth fund
	_, err = tx.Exec(`UPDATE growth_fund SET total_claimed = total_claimed + ?, last_claim_at = ? WHERE user_id = ?`,
		stage.RewardDM, now, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update growth fund"})
		return
	}

	// Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// Log transaction
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, 'growth_fund', ?, ?, ?)
	`, txID, playerID, stage.RewardDM, newBalance, stage.ID, fmt.Sprintf("成长基金阶段奖励: %s", stage.ID), now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{"balance": newBalance})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"stageId":    stage.ID,
		"dmClaimed":  stage.RewardDM,
		"newBalance": newBalance,
	})
}

// ─── Gift Packs ──────────────────────────────────────────────────────────────

// GetGiftPacks returns available gift packs for the player.
func (h *WalletHandler) GetGiftPacks(c *gin.Context) {
	playerID := c.GetString("user_id")

	packs := make([]gin.H, 0)
	for _, pack := range engine.GiftPacks {
		packInfo := gin.H{
			"id":       pack.ID,
			"name":     pack.Name,
			"costDM":   pack.CostDM,
			"contents": pack.Contents,
			"onceOnly": pack.OnceOnly,
		}

		// Check availability
		if pack.OnceOnly {
			var count int
			_ = h.db.QueryRow(`
				SELECT COUNT(*) FROM dark_matter_transactions
				WHERE user_id = ? AND type = 'gift_pack' AND ref_id = ?
			`, playerID, pack.ID).Scan(&count)
			packInfo["available"] = count == 0
		} else if pack.CooldownHours > 0 {
			var lastPurchase sql.NullString
			_ = h.db.QueryRow(`
				SELECT MAX(created_at) FROM dark_matter_transactions
				WHERE user_id = ? AND type = 'gift_pack' AND ref_id = ?
			`, playerID, pack.ID).Scan(&lastPurchase)
			if lastPurchase.Valid {
				t, _ := time.Parse(time.RFC3339, lastPurchase.String)
				cooldownEnd := t.Add(time.Duration(pack.CooldownHours) * time.Hour)
				packInfo["available"] = time.Now().After(cooldownEnd)
				packInfo["cooldownEnd"] = cooldownEnd.Format(time.RFC3339)
			} else {
				packInfo["available"] = true
			}
		} else {
			packInfo["available"] = true
		}

		packs = append(packs, packInfo)
	}

	c.JSON(http.StatusOK, gin.H{"packs": packs})
}

// BuyGiftPack purchases a gift pack.
func (h *WalletHandler) BuyGiftPack(c *gin.Context) {
	playerID := c.GetString("user_id")

	var req struct {
		PackID string `json:"packId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Find the pack
	var pack *engine.GiftPack
	for i := range engine.GiftPacks {
		if engine.GiftPacks[i].ID == req.PackID {
			pack = &engine.GiftPacks[i]
			break
		}
	}
	if pack == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown gift pack"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback()

	// Check balance
	var balance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query balance"})
		return
	}
	if balance < pack.CostDM {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient dark matter"})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	txID := uuid.New().String()

	// Deduct DM
	_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance - ? WHERE id = ?`, pack.CostDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct DM"})
		return
	}

	// Add consumption points
	_, err = tx.Exec(`UPDATE users SET consumption_points = consumption_points + ? WHERE id = ?`, pack.CostDM, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update points"})
		return
	}

	// Read new balance
	var newBalance int64
	err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
		return
	}

	// Log transaction
	_, err = tx.Exec(`
		INSERT INTO dark_matter_transactions (id, user_id, amount, balance_after, type, ref_id, description, created_at)
		VALUES (?, ?, ?, ?, 'gift_pack', ?, ?, ?)
	`, txID, playerID, -pack.CostDM, newBalance, pack.ID, fmt.Sprintf("购买限时礼包: %s", pack.Name), now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log transaction"})
		return
	}

	// Add resources to first planet (simplified — in production would need planet selection)
	if pack.Contents.Metal > 0 || pack.Contents.Crystal > 0 || pack.Contents.Deuterium > 0 {
		var planetID string
		err = tx.QueryRow(`SELECT id FROM planets WHERE owner_id = ? ORDER BY created_at ASC LIMIT 1`, playerID).Scan(&planetID)
		if err == nil {
			// Update planet resources directly in DB
			_, err = tx.Exec(`
				UPDATE planets SET
					metal = metal + ?,
					crystal = crystal + ?,
					deuterium = deuterium + ?
				WHERE id = ?
			`, pack.Contents.Metal, pack.Contents.Crystal, pack.Contents.Deuterium, planetID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add pack resources"})
				return
			}
		}
	}

	// Add DM from pack contents (if any)
	if pack.Contents.DarkMatter > 0 {
		_, err = tx.Exec(`UPDATE users SET dark_matter_balance = dark_matter_balance + ? WHERE id = ?`, pack.Contents.DarkMatter, playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add pack DM"})
			return
		}
		// Re-read balance
		err = tx.QueryRow(`SELECT dark_matter_balance FROM users WHERE id = ?`, playerID).Scan(&newBalance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read balance"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	if h.wsHub != nil {
		h.wsHub.SendTo(playerID, "dmBalanceChanged", gin.H{"balance": newBalance})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"packId":     pack.ID,
		"costDM":     pack.CostDM,
		"newBalance": newBalance,
		"contents":   pack.Contents,
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func parseIntParam(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
