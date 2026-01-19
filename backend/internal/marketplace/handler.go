package marketplace

import (
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"fmt"

	"github.com/gin-gonic/gin"
)

type MarketplaceHandler struct {
	service      *MarketplaceService
	auditService *audit.AuditService
}

func NewMarketplaceHandler(service *MarketplaceService, auditService *audit.AuditService) *MarketplaceHandler {
	return &MarketplaceHandler{service: service, auditService: auditService}
}

// GetAll handles getting all products
// @Summary Get all products
// @Description Get list of all products with pagination and filters (Admin)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status" Enums(active, inactive)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=ProductListResponse}
// @Failure 401 {object} utils.Response
// @Router /admin/products [get]
func (h *MarketplaceHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Security: If role is mahasiswa, force status to active
	role, _ := c.Get("role")
	if role == "mahasiswa" {
		status = "active"
	}

	params := ProductListParams{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	response, err := h.service.GetAllProducts(params)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve products", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", response)
}

// Get ByID handles getting product by ID
// @Summary Get product by ID
// @Description Get product details by ID (Admin)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} utils.Response{data=Product}
// @Failure 404 {object} utils.Response
// @Router /admin/products/{id} [get]
func (h *MarketplaceHandler) GetByID(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	product, err := h.service.GetProductByID(uint(productID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

// Create handles creating new product
// @Summary Create product
// @Description Create a new product (Admin only)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product details"
// @Success 201 {object} utils.Response{data=Product}
// @Failure 400 {object} utils.Response
// @Router /admin/products [post]
func (h *MarketplaceHandler) Create(c *gin.Context) {
	adminID := c.GetUint("user_id")

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	product, err := h.service.CreateProduct(&req, adminID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "CREATE_PRODUCT",
		Entity:    "PRODUCT",
		EntityID:  product.ID,
		Details:   "Admin created new product: " + product.Name,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Update handles updating product
// @Summary Update product
// @Description Update product information (Admin only)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body UpdateProductRequest true "Update data"
// @Success 200 {object} utils.Response{data=Product}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /admin/products/{id} [put]
func (h *MarketplaceHandler) Update(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(uint(productID), &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", product)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "UPDATE_PRODUCT",
		Entity:    "PRODUCT",
		EntityID:  product.ID,
		Details:   "Admin updated product: " + product.Name,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Delete handles deleting product
// @Summary Delete product
// @Description Delete product (Admin only)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /admin/products/{id} [delete]
func (h *MarketplaceHandler) Delete(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	if err := h.service.DeleteProduct(uint(productID)); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		utils.ErrorResponse(c, statusCode, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)

	// Log activity
	adminID := c.GetUint("user_id")
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    adminID,
		Action:    "DELETE_PRODUCT",
		Entity:    "PRODUCT",
		EntityID:  uint(productID),
		Details:   "Admin deleted product ID: " + strconv.FormatUint(productID, 10),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// Purchase handles product purchase
// @Summary Purchase product
// @Description Purchase a product with points
// @Tags Marketplace
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PurchaseRequest true "Purchase details"
// @Success 200 {object} utils.Response{data=MarketplaceTransaction}
// @Failure 400 {object} utils.Response
// @Router /marketplace/purchase [post]
func (h *MarketplaceHandler) Purchase(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	txn, err := h.service.PurchaseProduct(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Purchase successful", txn)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "PURCHASE_PRODUCT",
		Entity:    "MARKETPLACE_TRANSACTION",
		EntityID:  txn.ID,
		Details:   fmt.Sprintf("User purchased %d units of product ID %d", txn.Quantity, req.ProductID),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// GetTransactions handles getting all marketplace transactions
// @Summary Get marketplace transactions
// @Description Get list of all marketplace transactions (Admin)
// @Tags Admin - Marketplace
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /admin/marketplace/transactions [get]
func (h *MarketplaceHandler) GetTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, total, err := h.service.GetTransactions(limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", gin.H{
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}
