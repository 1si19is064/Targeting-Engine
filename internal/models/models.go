package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Campaign represents an advertising campaign
type Campaign struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ImageURL  string    `json:"img" db:"image_url"`
	CTA       string    `json:"cta" db:"cta"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CampaignResponse represents the response format for campaigns
type CampaignResponse struct {
	CID string `json:"cid"`
	IMG string `json:"img"`
	CTA string `json:"cta"`
}

// TargetingRule represents targeting criteria for campaigns
type TargetingRule struct {
	ID         int64       `json:"id" db:"id"`
	CampaignID string      `json:"campaign_id" db:"campaign_id"`
	Dimension  string      `json:"dimension" db:"dimension"` // "country", "os", "app"
	RuleType   string      `json:"rule_type" db:"rule_type"` // "include" or "exclude"
	Values     StringArray `json:"values" db:"values"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// StringArray custom type for handling array of strings in database
type StringArray []string

// Value implements driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}

	return json.Unmarshal(bytes, sa)
}

// DeliveryRequest represents the incoming request parameters
type DeliveryRequest struct {
	App     string `json:"app"`
	Country string `json:"country"`
	OS      string `json:"os"`
}

// ErrorResponse represents error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// Constants for campaign status
const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
)

// Constants for targeting dimensions
const (
	DimensionCountry = "country"
	DimensionOS      = "os"
	DimensionApp     = "app"
)

// Constants for rule types
const (
	RuleTypeInclude = "include"
	RuleTypeExclude = "exclude"
)

// Validate validates the delivery request
func (dr *DeliveryRequest) Validate() error {
	if dr.App == "" {
		return fmt.Errorf("missing app param")
	}
	if dr.Country == "" {
		return fmt.Errorf("missing country param")
	}
	if dr.OS == "" {
		return fmt.Errorf("missing os param")
	}
	return nil
}

// ToCampaignResponse converts Campaign to CampaignResponse
func (c *Campaign) ToCampaignResponse() CampaignResponse {
	return CampaignResponse{
		CID: c.ID,
		IMG: c.ImageURL,
		CTA: c.CTA,
	}
}
