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
func (h *MarketplaceHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

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

// GetByID handles getting product by ID
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
func (h *MarketplaceHandler) Purchase(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	err := h.service.PurchaseProduct(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Purchase successful", nil)

	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "PURCHASE_PRODUCT",
		Entity:    "PRODUCT",
		EntityID:  req.ProductID,
		Details:   fmt.Sprintf("User purchased units of product ID %d", req.ProductID),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// GetTransactions handles getting all marketplace transactions from consolidated wallet_transactions
func (h *MarketplaceHandler) GetTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	transactions, total, err := h.service.GetTransactions(limit, page)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Marketplace transactions retrieved", gin.H{
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"page":         page,
	})
}

func (h *MarketplaceHandler) GetCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartResponse, err := h.service.GetCart(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Keranjang berhasil diambil", gin.H{
		"items":       cartResponse.Items,
		"total_price": cartResponse.TotalPrice,
	})
}

func (h *MarketplaceHandler) AddToCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	fmt.Printf("DEBUG: Adding to cart - UserID: %d, ProductID: %d, Quantity: %d\n", userID, req.ProductID, req.Quantity)

	if err := h.service.AddToCart(userID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Produk berhasil ditambahkan ke keranjang", nil)
}

func (h *MarketplaceHandler) UpdateCartItem(c *gin.Context) {
	userID := c.GetUint("user_id")
	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.UpdateCartItem(userID, uint(itemID), req.Quantity); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Keranjang berhasil diperbarui", nil)
}

func (h *MarketplaceHandler) RemoveFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.service.RemoveFromCart(userID, uint(itemID)); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Produk berhasil dihapus dari keranjang", nil)
}

func (h *MarketplaceHandler) Checkout(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req CartCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	if err := h.service.Checkout(userID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Checkout berhasil!", nil)

	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID,
		Action:    "CART_CHECKOUT",
		Entity:    "WALLET",
		EntityID:  userID,
		Details:   "User completed checkout from cart",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}
