package http

import (
	"net/http"
	"strconv"

	"day5/internal/application/usecases"
	"day5/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles HTTP requests for product operations
// This is the interface/presentation layer in Clean Architecture
type ProductHandler struct {
	productUseCase *usecases.ProductUseCase
}

// NewProductHandler creates a new product handler with dependency injection
func NewProductHandler(productUseCase *usecases.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

// ProductResponse represents the HTTP response for product operations
type ProductResponse struct {
	ID          string  `json:"id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Message     string  `json:"message,omitempty"`
}

// ProductListResponse represents the response for listing products
type ProductListResponse struct {
	Products []*ProductResponse `json:"products"`
	Count    int                `json:"count"`
	Message  string             `json:"message,omitempty"`
}

// CreateProduct handles POST /api/v1/product
// @Summary Create a new product
// @Description Creates a new product with the provided details
// @Tags Products
// @Accept json
// @Produce json
// @Param product body usecases.CreateProductRequest true "Product details"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/product [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req usecases.CreateProductRequest

	// Bind and validate request
	// The binding:"required" tag ensures required fields are present
	// The binding:"gt=0" tag ensures numeric values are greater than 0
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Call use case (business logic layer)
	product, err := h.productUseCase.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.entityToResponse(product, "Product successfully created")
	c.JSON(http.StatusCreated, response)
}

// GetProduct handles GET /api/v1/product/:id
// @Summary Get a product by ID
// @Description Retrieves a product by its unique identifier
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/product/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product ID is required",
		})
		return
	}

	product, err := h.productUseCase.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	response := h.entityToResponse(product, "")
	c.JSON(http.StatusOK, response)
}

// GetProducts handles GET /api/v1/products
// @Summary List all products
// @Description Retrieves a list of all products with optional pagination
// @Tags Products
// @Produce json
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} ProductListResponse
// @Failure 500 {object} map[string]any
// @Router /api/v1/products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.productUseCase.GetAllProducts(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve products",
			"details": err.Error(),
		})
		return
	}

	// Convert entities to response format
	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = h.entityToResponse(product, "")
	}

	response := &ProductListResponse{
		Products: productResponses,
		Count:    len(productResponses),
		Message:  "Products retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProduct handles PUT /api/v1/product/:id
// @Summary Update a product
// @Description Updates product price and/or quantity
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body usecases.UpdateProductRequest true "Product update details"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/product/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product ID is required",
		})
		return
	}

	var req usecases.UpdateProductRequest

	// Bind and validate request
	// The omitempty tag means the field is optional
	// The binding:"omitempty,gt=0" tag validates only if the field is provided
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Call use case
	product, err := h.productUseCase.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update product",
			"details": err.Error(),
		})
		return
	}

	response := h.entityToResponse(product, "Product successfully updated")
	c.JSON(http.StatusOK, response)
}

// SearchProducts handles GET /api/v1/products/search
// @Summary Search products by name
// @Description Searches for products containing the specified name
// @Tags Products
// @Produce json
// @Param name query string true "Product name to search for"
// @Success 200 {object} ProductListResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/products/search [get]
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search name is required",
		})
		return
	}

	products, err := h.productUseCase.SearchProducts(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search products",
			"details": err.Error(),
		})
		return
	}

	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = h.entityToResponse(product, "")
	}

	response := &ProductListResponse{
		Products: productResponses,
		Count:    len(productResponses),
		Message:  "Products found",
	}

	c.JSON(http.StatusOK, response)
}

// GetAvailableProducts handles GET /api/v1/products/available
// @Summary Get available products
// @Description Retrieves products that have quantity > 0
// @Tags Products
// @Produce json
// @Success 200 {object} ProductListResponse
// @Failure 500 {object} map[string]any
// @Router /api/v1/products/available [get]
func (h *ProductHandler) GetAvailableProducts(c *gin.Context) {
	products, err := h.productUseCase.GetAvailableProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve available products",
			"details": err.Error(),
		})
		return
	}

	productResponses := make([]*ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = h.entityToResponse(product, "")
	}

	response := &ProductListResponse{
		Products: productResponses,
		Count:    len(productResponses),
		Message:  "Available products retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

// Helper method to convert domain entity to HTTP response
func (h *ProductHandler) entityToResponse(product *entities.Product, message string) *ProductResponse {
	return &ProductResponse{
		ID:          product.ID,
		ProductName: product.ProductName,
		Price:       product.Price,
		Quantity:    product.Quantity,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Message:     message,
	}
}
