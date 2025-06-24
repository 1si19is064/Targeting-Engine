package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"Targeting-Engine/internal/models"
	"Targeting-Engine/internal/services"
	"Targeting-Engine/internal/utils"
)

type DeliveryController struct {
	targetingService *services.TargetingService
}

func NewDeliveryController(targetingService *services.TargetingService) *DeliveryController {
	return &DeliveryController{
		targetingService: targetingService,
	}
}

// GetCampaigns handles the delivery endpoint
func (dc *DeliveryController) GetCampaigns(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Parse query parameters
	req := &models.DeliveryRequest{
		App:     r.URL.Query().Get("app"),
		Country: r.URL.Query().Get("country"),
		OS:      r.URL.Query().Get("os"),
	}

	// Validate request
	if err := req.Validate(); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get matching campaigns
	campaigns, err := dc.targetingService.GetMatchingCampaigns(req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Handle empty result
	if len(campaigns) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Convert to response format
	var response []models.CampaignResponse
	for _, campaign := range campaigns {
		response = append(response, campaign.ToCampaignResponse())
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error but don't change response since headers are already written
		utils.LogError("Failed to encode response", err)
	}

	// Log request duration for monitoring
	duration := time.Since(startTime)
	utils.LogRequest(r, http.StatusOK, duration)
}