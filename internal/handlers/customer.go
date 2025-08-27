package handlers

import (
	"net/http"

	"day5/internal/database"
	"day5/internal/models"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct{}

func NewCustomerHandler() *CustomerHandler {
	return &CustomerHandler{}
}

// CreateCustomer handles POST /customer
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req models.CustomerRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Create customer model from request
	customer := models.Customer{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}

	// Save to database
	if err := database.GetDB().Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create customer",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	response := customer.ToResponse("customer successfully created")
	c.JSON(http.StatusCreated, response)
}

// GetCustomers handles GET /customers
func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	var customers []models.Customer

	if err := database.GetDB().Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch customers",
		})
		return
	}

	var responses []models.CustomerResponse
	for _, customer := range customers {
		responses = append(responses, customer.ToResponse(""))
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": responses,
		"count":     len(responses),
	})
}

// GetCustomer handles GET /customer/:id
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer

	if err := database.GetDB().First(&customer, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Customer not found",
		})
		return
	}

	response := customer.ToResponse("")
	c.JSON(http.StatusOK, response)
}
