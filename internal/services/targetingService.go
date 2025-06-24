package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"Targeting-Engine/internal/models"
)

type TargetingService struct {
	db               *sql.DB
	campaignCache    map[string]*models.Campaign
	rulesCache       map[string][]*models.TargetingRule
	cacheMutex       sync.RWMutex
	lastCacheUpdate  time.Time
	cacheRefreshRate time.Duration
}

func NewTargetingService(db *sql.DB) *TargetingService {
	service := &TargetingService{
		db:               db,
		campaignCache:    make(map[string]*models.Campaign),
		rulesCache:       make(map[string][]*models.TargetingRule),
		cacheRefreshRate: 30 * time.Second, // Refresh cache every 30 seconds
	}

	// Initial cache load
	service.refreshCache()

	// Start background cache refresh
	go service.startCacheRefresh()

	return service
}

// GetMatchingCampaigns returns campaigns that match the delivery request
func (ts *TargetingService) GetMatchingCampaigns(req *models.DeliveryRequest) ([]*models.Campaign, error) {
	// Ensure cache is fresh
	ts.ensureCacheIsFresh()

	ts.cacheMutex.RLock()
	defer ts.cacheMutex.RUnlock()

	var matchingCampaigns []*models.Campaign

	// Iterate through all campaigns
	for campaignID, campaign := range ts.campaignCache {
		// Skip inactive campaigns
		if campaign.Status != models.StatusActive {
			continue
		}

		// Check if campaign matches targeting rules
		if ts.matchesTargetingRules(campaignID, req) {
			matchingCampaigns = append(matchingCampaigns, campaign)
		}
	}

	return matchingCampaigns, nil
}

// matchesTargetingRules checks if a request matches all targeting rules for a campaign
func (ts *TargetingService) matchesTargetingRules(campaignID string, req *models.DeliveryRequest) bool {
	rules, exists := ts.rulesCache[campaignID]
	if !exists {
		// No rules means campaign matches all requests
		return true
	}

	// Group rules by dimension
	dimensionRules := make(map[string][]*models.TargetingRule)
	for _, rule := range rules {
		dimensionRules[rule.Dimension] = append(dimensionRules[rule.Dimension], rule)
	}

	// Check each dimension
	for dimension, dimRules := range dimensionRules {
		if !ts.matchesDimensionRules(dimension, dimRules, req) {
			return false
		}
	}

	return true
}

// matchesDimensionRules checks if request matches rules for a specific dimension
func (ts *TargetingService) matchesDimensionRules(dimension string, rules []*models.TargetingRule, req *models.DeliveryRequest) bool {
	var requestValue string
	switch dimension {
	case models.DimensionCountry:
		requestValue = strings.ToLower(req.Country)
	case models.DimensionOS:
		requestValue = strings.ToLower(req.OS)
	case models.DimensionApp:
		requestValue = req.App
	default:
		return false
	}

	var includeRules, excludeRules []*models.TargetingRule
	for _, rule := range rules {
		if rule.RuleType == models.RuleTypeInclude {
			includeRules = append(includeRules, rule)
		} else if rule.RuleType == models.RuleTypeExclude {
			excludeRules = append(excludeRules, rule)
		}
	}

	// Check exclude rules first
	for _, rule := range excludeRules {
		if ts.valueInList(requestValue, rule.Values) {
			return false // Excluded, doesn't match
		}
	}

	// Check include rules
	if len(includeRules) == 0 {
		return true // No include rules means all values are included
	}

	for _, rule := range includeRules {
		if ts.valueInList(requestValue, rule.Values) {
			return true // Found in include list
		}
	}

	return false // Not found in any include list
}

// valueInList checks if a value exists in a list (case-insensitive for country/os)
func (ts *TargetingService) valueInList(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(value, item) {
			return true
		}
	}
	return false
}

// refreshCache loads campaigns and rules from database into memory
func (ts *TargetingService) refreshCache() {
	ts.cacheMutex.Lock()
	defer ts.cacheMutex.Unlock()

	// Load campaigns
	campaigns, err := ts.loadCampaigns()
	if err != nil {
		log.Printf("Error loading campaigns: %v", err)
		return
	}

	// Load targeting rules
	rules, err := ts.loadTargetingRules()
	if err != nil {
		log.Printf("Error loading targeting rules: %v", err)
		return
	}

	// Update cache
	ts.campaignCache = campaigns
	ts.rulesCache = rules
	ts.lastCacheUpdate = time.Now()

	log.Printf("Cache refreshed successfully. Campaigns: %d, Rules: %d",
		len(campaigns), len(rules))
}

// loadCampaigns loads all campaigns from database
func (ts *TargetingService) loadCampaigns() (map[string]*models.Campaign, error) {
	query := `
		SELECT id, name, image_url, cta, status, created_at, updated_at 
		FROM campaigns
	`

	rows, err := ts.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	campaigns := make(map[string]*models.Campaign)
	for rows.Next() {
		var campaign models.Campaign
		err := rows.Scan(
			&campaign.ID,
			&campaign.Name,
			&campaign.ImageURL,
			&campaign.CTA,
			&campaign.Status,
			&campaign.CreatedAt,
			&campaign.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns[campaign.ID] = &campaign
	}

	return campaigns, nil
}

// loadTargetingRules loads all targeting rules from database
func (ts *TargetingService) loadTargetingRules() (map[string][]*models.TargetingRule, error) {
	query := `
		SELECT id, campaign_id, dimension, rule_type, values, created_at, updated_at
		FROM targeting_rules
	`

	rows, err := ts.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query targeting rules: %w", err)
	}
	defer rows.Close()

	rules := make(map[string][]*models.TargetingRule)
	for rows.Next() {
		var rule models.TargetingRule
		err := rows.Scan(
			&rule.ID,
			&rule.CampaignID,
			&rule.Dimension,
			&rule.RuleType,
			&rule.Values,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan targeting rule: %w", err)
		}
		rules[rule.CampaignID] = append(rules[rule.CampaignID], &rule)
	}

	return rules, nil
}

// startCacheRefresh starts background cache refresh
func (ts *TargetingService) startCacheRefresh() {
	ticker := time.NewTicker(ts.cacheRefreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ts.refreshCache()
		}
	}
}

// ensureCacheIsFresh refreshes cache if it's stale
func (ts *TargetingService) ensureCacheIsFresh() {
	if time.Since(ts.lastCacheUpdate) > ts.cacheRefreshRate {
		go ts.refreshCache()
	}
}
