package external

import (
	"net/http"
	"strconv"
	"wallet-point/internal/audit"
	"wallet-point/internal/marketplace" // Add this
	"wallet-point/internal/mission"     // Add this
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service      *Service
	auditService *audit.AuditService // Add this
}

func NewHandler(service *Service, auditService *audit.AuditService) *Handler {
	return &Handler{
		service:      service,
		auditService: auditService, // Add this
	}
}

// @Summary List external point sources
// @Tags External
// @Produce json
// @Success 200 {object} utils.Response
// @Router /external/sources [get]
func (h *Handler) ListSources(c *gin.Context) {
	sources, err := h.service.ListSources()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch sources", nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Sources fetched successfully", sources)
}

// @Summary Sync points from external source
// @Tags External
// @Accept json
// @Produce json
// @Param request body SyncRequest true "Sync Request"
// @Success 200 {object} utils.Response
// @Router /mahasiswa/external/sync [post]
func (h *Handler) SyncPoints(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	log, err := h.service.SyncPoints(userID.(uint), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Points synced successfully", log)

	// Log activity
	h.auditService.LogActivity(audit.CreateAuditParams{
		UserID:    userID.(uint),
		Action:    "EXTERNAL_SYNC",
		Entity:    "EXTERNAL_LOG",
		EntityID:  log.ID,
		Details:   "User synced points from external source",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
}

// @Summary Add external source (Admin only)
// @Tags External
// @Accept json
// @Produce json
// @Param request body ExternalSourceCreateRequest true "Create Source Request"
// @Success 201 {object} utils.Response
// @Router /admin/external/sources [post]
func (h *Handler) RegisterSource(c *gin.Context) {
	var req ExternalSourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	source, err := h.service.RegisterSource(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register source", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Source registered successfully", source)
}

// @Summary Add marketplace product from external source
// @Tags External
// @Accept json
// @Produce json
// @Param source_id query int true "Source ID"
// @Param request body marketplace.CreateProductRequest true "Product Details"
// @Success 201 {object} utils.Response
// @Router /external/marketplace/products [post]
func (h *Handler) RegisterProduct(c *gin.Context) {
	sourceIDStr := c.Query("source_id")
	sourceID, _ := strconv.ParseUint(sourceIDStr, 10, 32)

	var req marketplace.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	product, err := h.service.CreateProductFromExternal(uint(sourceID), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "External product registered", product)
}

// @Summary Add mission from external source
// @Tags External
// @Accept json
// @Produce json
// @Param source_id query int true "Source ID"
// @Param request body mission.CreateMissionRequest true "Mission Details"
// @Success 201 {object} utils.Response
// @Router /external/missions [post]
func (h *Handler) RegisterMission(c *gin.Context) {
	sourceIDStr := c.Query("source_id")
	sourceID, _ := strconv.ParseUint(sourceIDStr, 10, 32)

	var req mission.CreateMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	mission, err := h.service.CreateMissionFromExternal(uint(sourceID), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "External mission registered", mission)
}
