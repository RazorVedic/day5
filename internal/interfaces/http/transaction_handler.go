package http

import (
	"net/http"
	"strconv"
	"time"

	"day5/internal/application/usecases"
	"day5/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// TransactionHandler handles HTTP requests for transaction and analytics operations
type TransactionHandler struct {
	transactionUseCase *usecases.TransactionUseCase
}

// NewTransactionHandler creates a new transaction handler with dependency injection
func NewTransactionHandler(transactionUseCase *usecases.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{
		transactionUseCase: transactionUseCase,
	}
}

// TransactionResponse represents the HTTP response for transaction operations
type TransactionResponse struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	CustomerID    string  `json:"customer_id"`
	CustomerName  string  `json:"customer_name,omitempty"`
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name,omitempty"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	Description   string  `json:"description"`
	TransactionAt string  `json:"transaction_at"`
	CreatedAt     string  `json:"created_at"`
}

// TransactionHistoryResponse represents the response for transaction history
type TransactionHistoryResponse struct {
	Transactions []*TransactionResponse `json:"transactions"`
	Count        int                    `json:"count"`
	TotalAmount  float64                `json:"total_amount"`
	Message      string                 `json:"message,omitempty"`
}

// GetTransactionHistory handles GET /api/v1/transactions
// @Summary Get transaction history
// @Description Retrieves transaction history with optional filters
// @Tags Transactions
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param product_id query string false "Filter by product ID"
// @Param type query string false "Filter by transaction type (order, refund, credit)"
// @Param start_date query string false "Start date filter (RFC3339 format)"
// @Param end_date query string false "End date filter (RFC3339 format)"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} TransactionHistoryResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	// Parse query parameters with validation
	filters := usecases.TransactionFilters{
		CustomerID: c.Query("customer_id"),
		ProductID:  c.Query("product_id"),
		Type:       c.Query("type"),
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filters.Limit = limit
	filters.Offset = offset

	// Parse date filters if provided
	// RFC3339 format: "2006-01-02T15:04:05Z07:00"
	if startDate := c.Query("start_date"); startDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &parsedDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid start_date format",
				"details": "Use RFC3339 format: 2006-01-02T15:04:05Z07:00",
			})
			return
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &parsedDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid end_date format",
				"details": "Use RFC3339 format: 2006-01-02T15:04:05Z07:00",
			})
			return
		}
	}

	// Call use case
	transactions, err := h.transactionUseCase.GetTransactionHistory(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve transaction history",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format and calculate total
	transactionResponses := make([]*TransactionResponse, len(transactions))
	totalAmount := 0.0
	for i, transaction := range transactions {
		transactionResponses[i] = h.entityToResponse(transaction)
		totalAmount += transaction.GetRevenueAmount()
	}

	response := &TransactionHistoryResponse{
		Transactions: transactionResponses,
		Count:        len(transactionResponses),
		TotalAmount:  totalAmount,
		Message:      "Transaction history retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionStats handles GET /api/v1/transactions/stats
// @Summary Get business statistics
// @Description Retrieves comprehensive business analytics and statistics
// @Tags Transactions
// @Produce json
// @Param period query string false "Statistics period (today, this_week, this_month, all_time)" default("all_time")
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/transactions/stats [get]
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	periodParam := c.DefaultQuery("period", "all_time")

	// Validate period parameter
	var period usecases.StatsPeriod
	switch periodParam {
	case "today":
		period = usecases.StatsPeriodToday
	case "this_week":
		period = usecases.StatsPeriodThisWeek
	case "this_month":
		period = usecases.StatsPeriodThisMonth
	case "all_time":
		period = usecases.StatsPeriodAllTime
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Invalid period parameter",
			"valid_periods": []string{"today", "this_week", "this_month", "all_time"},
		})
		return
	}

	stats, err := h.transactionUseCase.GetBusinessStats(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve business statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetComprehensiveStats handles GET /api/v1/transactions/stats/comprehensive
// @Summary Get comprehensive statistics for all periods
// @Description Retrieves business statistics for all time periods (today, week, month, all-time)
// @Tags Transactions
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/transactions/stats/comprehensive [get]
func (h *TransactionHandler) GetComprehensiveStats(c *gin.Context) {
	stats, err := h.transactionUseCase.GetComprehensiveStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve comprehensive statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCustomerTransactionSummary handles GET /api/v1/transactions/customer/:customer_id/summary
// @Summary Get customer transaction summary
// @Description Retrieves transaction summary for a specific customer
// @Tags Transactions
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/transactions/customer/{customer_id}/summary [get]
func (h *TransactionHandler) GetCustomerTransactionSummary(c *gin.Context) {
	customerID := c.Param("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Customer ID is required",
		})
		return
	}

	summary, err := h.transactionUseCase.GetCustomerTransactionSummary(c.Request.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve customer transaction summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetRevenueAnalytics handles GET /api/v1/transactions/revenue/analytics
// @Summary Get revenue analytics
// @Description Retrieves detailed revenue analytics including daily/monthly trends
// @Tags Transactions
// @Produce json
// @Param days query int false "Number of days for daily revenue analysis" default(30)
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/transactions/revenue/analytics [get]
func (h *TransactionHandler) GetRevenueAnalytics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	analytics, err := h.transactionUseCase.GetRevenueAnalytics(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve revenue analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// Helper method to convert domain entity to HTTP response
func (h *TransactionHandler) entityToResponse(transaction *entities.Transaction) *TransactionResponse {
	response := &TransactionResponse{
		ID:            transaction.ID,
		OrderID:       transaction.OrderID,
		CustomerID:    transaction.CustomerID,
		ProductID:     transaction.ProductID,
		Type:          string(transaction.Type),
		Amount:        transaction.Amount,
		Quantity:      transaction.Quantity,
		UnitPrice:     transaction.UnitPrice,
		Description:   transaction.Description,
		TransactionAt: transaction.TransactionAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:     transaction.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add related entity information if available
	if transaction.Customer != nil {
		response.CustomerName = transaction.Customer.Name
	}
	if transaction.Product != nil {
		response.ProductName = transaction.Product.ProductName
	}

	return response
}
