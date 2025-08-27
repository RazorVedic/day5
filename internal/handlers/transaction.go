package handlers

import (
	"net/http"
	"strconv"
	"time"

	"day5/internal/database"
	"day5/internal/models"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct{}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

// GetTransactionHistory handles GET /transactions - for retailer to view all business transactions
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	db := database.GetDB()

	// Parse query parameters for filtering
	limit := 100 // default limit
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Date filtering
	var startDate, endDate time.Time
	query := db.Model(&models.Transaction{}).
		Preload("Customer").
		Preload("Product").
		Preload("Order")

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
			query = query.Where("created_at >= ?", startDate)
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed.Add(24 * time.Hour) // Include the entire end date
			query = query.Where("created_at < ?", endDate)
		}
	}

	// Transaction type filtering
	if txType := c.Query("type"); txType != "" {
		query = query.Where("type = ?", txType)
	}

	// Customer filtering
	if customerID := c.Query("customer_id"); customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}

	// Product filtering
	if productID := c.Query("product_id"); productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	var transactions []models.Transaction
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch transaction history",
		})
		return
	}

	// Calculate total amount for the filtered transactions
	var totalAmount float64
	for _, transaction := range transactions {
		totalAmount += transaction.Amount
	}

	var responses []models.TransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, transaction.ToResponse())
	}

	c.JSON(http.StatusOK, models.TransactionHistoryResponse{
		Transactions: responses,
		Count:        len(responses),
		TotalAmount:  totalAmount,
	})
}

// GetTransactionStats handles GET /transactions/stats - for retailer dashboard
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	db := database.GetDB()

	// Calculate stats for today, this week, this month
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type StatsResult struct {
		TotalAmount float64 `json:"total_amount"`
		OrderCount  int64   `json:"order_count"`
	}

	var todayStats, weekStats, monthStats, allTimeStats StatsResult

	// Today's stats
	db.Model(&models.Transaction{}).
		Where("created_at >= ? AND type = ?", today, models.TransactionTypeOrder).
		Select("COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as order_count").
		Scan(&todayStats)

	// This week's stats
	db.Model(&models.Transaction{}).
		Where("created_at >= ? AND type = ?", weekStart, models.TransactionTypeOrder).
		Select("COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as order_count").
		Scan(&weekStats)

	// This month's stats
	db.Model(&models.Transaction{}).
		Where("created_at >= ? AND type = ?", monthStart, models.TransactionTypeOrder).
		Select("COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as order_count").
		Scan(&monthStats)

	// All time stats
	db.Model(&models.Transaction{}).
		Where("type = ?", models.TransactionTypeOrder).
		Select("COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as order_count").
		Scan(&allTimeStats)

	// Top selling products
	type ProductSales struct {
		ProductID   string  `json:"product_id"`
		ProductName string  `json:"product_name"`
		TotalSold   int64   `json:"total_sold"`
		Revenue     float64 `json:"revenue"`
	}

	var topProducts []ProductSales
	db.Table("transactions t").
		Select("t.product_id, p.product_name, SUM(t.quantity) as total_sold, SUM(t.amount) as revenue").
		Joins("JOIN products p ON t.product_id = p.id").
		Where("t.type = ?", models.TransactionTypeOrder).
		Group("t.product_id, p.product_name").
		Order("revenue DESC").
		Limit(5).
		Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{
		"today":        todayStats,
		"week":         weekStats,
		"month":        monthStats,
		"all_time":     allTimeStats,
		"top_products": topProducts,
		"stats_date":   now.Format("2006-01-02 15:04:05"),
	})
}
