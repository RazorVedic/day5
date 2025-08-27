package handlers

import (
	"net/http"

	"day5/internal/database"
	"day5/internal/models"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// CreateProduct handles POST /product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.ProductRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Create product model from request
	product := models.Product{
		ProductName: req.ProductName,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	// Save to database
	if err := database.GetDB().Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	response := product.ToResponse("product successfully added")
	c.JSON(http.StatusCreated, response)
}

// GetProducts handles GET /products (bonus endpoint for testing)
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var products []models.Product

	if err := database.GetDB().Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// GetProduct handles GET /product/:id (bonus endpoint for testing)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := database.GetDB().First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct handles PUT /product/:id - for retailer to update price and quantity
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req models.ProductUpdateRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find existing product
	var product models.Product
	if err := database.GetDB().First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	// Update fields if provided
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Quantity != nil {
		product.Quantity = *req.Quantity
	}

	// Save changes
	if err := database.GetDB().Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update product",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	response := product.ToResponse("Product successfully updated")
	c.JSON(http.StatusOK, response)
}
