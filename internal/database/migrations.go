package database

import (
	"database/sql"
	"fmt"
)

// Migrate runs database migrations
func Migrate(db *sql.DB) error {
	migrations := []string{
		createCampaignsTable,
		createTargetingRulesTable,
		insertSampleData,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	fmt.Println("Database migrations completed successfully")
	return nil
}

const createCampaignsTable = `
CREATE TABLE IF NOT EXISTS campaigns (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    cta VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
`

const createTargetingRulesTable = `
CREATE TABLE IF NOT EXISTS targeting_rules (
    id SERIAL PRIMARY KEY,
    campaign_id VARCHAR(255) NOT NULL,
    dimension VARCHAR(50) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    values JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_targeting_rules_campaign_id ON targeting_rules(campaign_id);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_dimension ON targeting_rules(dimension);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_rule_type ON targeting_rules(rule_type);
`

const insertSampleData = `
-- Insert sample campaigns
INSERT INTO campaigns (id, name, image_url, cta, status) VALUES
('spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', 'ACTIVE'),
('duolingo', 'Duolingo: Best way to learn', 'https://somelink2', 'Install', 'ACTIVE'),
('subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

-- Insert sample targeting rules
INSERT INTO targeting_rules (campaign_id, dimension, rule_type, values) VALUES
('spotify', 'country', 'include', '["US", "Canada"]'),
('duolingo', 'os', 'include', '["Android", "iOS"]'),
('duolingo', 'country', 'exclude', '["US"]'),
('subwaysurfer', 'os', 'include', '["Android"]'),
('subwaysurfer', 'app', 'include', '["com.gametion.ludokinggame"]')
ON CONFLICT DO NOTHING;
`
