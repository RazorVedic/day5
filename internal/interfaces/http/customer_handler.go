package http

import (
	"net/http"
	"strconv"

	"day5/internal/application/usecases"
	"day5/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	customerUseCase *usecases.CustomerUseCase
}

// NewCustomerHandler creates a new customer handler with dependency injection
func NewCustomerHandler(customerUseCase *usecases.CustomerUseCase) *CustomerHandler {
	return &CustomerHandler{
		customerUseCase: customerUseCase,
	}
}

// CustomerResponse represents the HTTP response for customer operations
type CustomerResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Message   string `json:"message,omitempty"`
}

// CustomerListResponse represents the response for listing customers
type CustomerListResponse struct {
	Customers []*CustomerResponse `json:"customers"`
	Count     int                 `json:"count"`
	Message   string              `json:"message,omitempty"`
}

// CreateCustomer handles POST /api/v1/customer
// @Summary Create a new customer
// @Description Registers a new customer with validation
// @Tags Customers
// @Accept json
// @Produce json
// @Param customer body usecases.CreateCustomerRequest true "Customer details"
// @Success 201 {object} CustomerResponse
// @Failure 400 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/customer [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req usecases.CreateCustomerRequest

	// Struct tags explanation:
	// - binding:"required" ensures the field must be present
	// - binding:"email" validates email format using built-in validator
	// - json:"name" specifies the JSON field name for marshaling/unmarshaling
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Call use case (business logic layer)
	customer, err := h.customerUseCase.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		// Handle business logic errors
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Customer with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create customer",
			"details": err.Error(),
		})
		return
	}

	response := h.entityToResponse(customer, "Customer successfully created")
	c.JSON(http.StatusCreated, response)
}

// GetCustomer handles GET /api/v1/customer/:id
// @Summary Get a customer by ID
// @Description Retrieves a customer by their unique identifier
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} CustomerResponse
// @Failure 404 {object} map[string]any
// @Router /api/v1/customer/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Customer ID is required",
		})
		return
	}

	customer, err := h.customerUseCase.GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Customer not found",
		})
		return
	}

	response := h.entityToResponse(customer, "")
	c.JSON(http.StatusOK, response)
}

// GetCustomers handles GET /api/v1/customers
// @Summary List all customers
// @Description Retrieves a list of all customers with pagination
// @Tags Customers
// @Produce json
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} CustomerListResponse
// @Failure 500 {object} map[string]any
// @Router /api/v1/customers [get]
func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	customers, err := h.customerUseCase.GetAllCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve customers",
			"details": err.Error(),
		})
		return
	}

	customerResponses := make([]*CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = h.entityToResponse(customer, "")
	}

	response := &CustomerListResponse{
		Customers: customerResponses,
		Count:     len(customerResponses),
		Message:   "Customers retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// GetCooldownStatus handles GET /api/v1/customer/:id/cooldown
// @Summary Get customer cooldown status
// @Description Retrieves the cooldown status for order placement
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/customer/{id}/cooldown [get]
func (h *CustomerHandler) GetCooldownStatus(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Customer ID is required",
		})
		return
	}

	status, err := h.customerUseCase.GetCooldownStatus(c.Request.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cooldown status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SearchCustomers handles GET /api/v1/customers/search
// @Summary Search customers by name
// @Description Searches for customers containing the specified name
// @Tags Customers
// @Produce json
// @Param name query string true "Customer name to search for"
// @Success 200 {object} CustomerListResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/customers/search [get]
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search name is required",
		})
		return
	}

	customers, err := h.customerUseCase.SearchCustomers(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search customers",
			"details": err.Error(),
		})
		return
	}

	customerResponses := make([]*CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = h.entityToResponse(customer, "")
	}

	response := &CustomerListResponse{
		Customers: customerResponses,
		Count:     len(customerResponses),
		Message:   "Customers found",
	}

	c.JSON(http.StatusOK, response)
}

// Helper method to convert domain entity to HTTP response
func (h *CustomerHandler) entityToResponse(customer *entities.Customer, message string) *CustomerResponse {
	return &CustomerResponse{
		ID:        customer.ID,
		Name:      customer.Name,
		Email:     customer.Email,
		Phone:     customer.Phone,
		CreatedAt: customer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: customer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Message:   message,
	}
}
