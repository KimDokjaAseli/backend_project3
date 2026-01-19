package transfer

import (
	"fmt"
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for transfers
type Handler struct {
	service      *Service
	auditService *audit.AuditService
}

// NewHandler creates a new transfer handler
func NewHandler(service *Service, auditService *audit.AuditService) *Handler {
	return &Handler{service: service, auditService: auditService}
}

// CreateTransfer handles POST /transfer
// @Summary Create a new transfer
// @Description Transfer points from current user to another user
// @Tags Transfer
// @Accept json
// @Produce json
// @Param transfer body TransferRequest true "Transfer details"
// @Success 200 {object} TransferResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /transfer [post]
func (h *Handler) CreateTransfer(c *gin.Context) {
	// Get current user ID from JWT
	senderUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Create transfer
	transfer, err := h.service.CreateTransfer(senderUserID.(uint), req.ReceiverUserID, req.Amount, req.Description)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer completed successfully", gin.H{
		"transfer": transfer,
	})

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    senderUserID.(uint),
		Action:    "TRANSFER_POINTS",
		Entity:    "TRANSFER",
		EntityID:  transfer.ID,
		Details:   fmt.Sprintf("Transferred %d points to user %d", req.Amount, req.ReceiverUserID),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// GetMyTransfers handles GET /transfer/history
// @Summary Get user's transfer history
// @Description Get all transfers (sent and received) for the current user
// @Tags Transfer
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /transfer/history [get]
func (h *Handler) GetMyTransfers(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transfers, total, err := h.service.GetUserTransfers(userID.(uint), limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer history retrieved successfully", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetSentTransfers handles GET /transfer/sent
// @Summary Get sent transfers
// @Description Get all transfers sent by the current user
// @Tags Transfer
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /transfer/sent [get]
func (h *Handler) GetSentTransfers(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transfers, total, err := h.service.GetSentTransfers(userID.(uint), limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sent transfers retrieved successfully", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetReceivedTransfers handles GET /transfer/received
// @Summary Get received transfers
// @Description Get all transfers received by the current user
// @Tags Transfer
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /transfer/received [get]
func (h *Handler) GetReceivedTransfers(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transfers, total, err := h.service.GetReceivedTransfers(userID.(uint), limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Received transfers retrieved successfully", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetAllTransfers handles GET /admin/transfers (Admin only)
// @Summary Get all transfers
// @Description Admin endpoint to get all transfers in the system
// @Tags Admin
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /admin/transfers [get]
func (h *Handler) GetAllTransfers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transfers, total, err := h.service.GetAllTransfers(limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All transfers retrieved successfully", gin.H{
		"transfers": transfers,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetRecipientInfo handles GET /transfer/recipient/:id
// @Summary Get recipient info for verification
// @Description Find recipient name and role by ID
// @Tags Transfer
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} RecipientSummary
// @Security BearerAuth
// @Router /mahasiswa/transfer/recipient/{id} [get]
func (h *Handler) GetRecipientInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	recipient, err := h.service.FindRecipient(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recipient found", recipient)
}
