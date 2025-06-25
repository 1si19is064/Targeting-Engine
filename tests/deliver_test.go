package tests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"Targeting-Engine/internal/controllers"
	"Targeting-Engine/internal/database"
	"Targeting-Engine/internal/models"
	"Targeting-Engine/internal/services"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Get database URL from environment variable or use default
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		// Default test database URL - adjust as needed
		testDBURL = "postgresql://postgres.wywkucanulrrkqgexwcp:Tejas%402001@aws-0-ap-south-1.pooler.supabase.com:5432/postgres"
	}

	db, err := database.Connect(testDBURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: database ping failed: %v", err)
		return nil
	}

	// Clean up tables
	_, err = db.Exec("TRUNCATE TABLE targeting_rules, campaigns CASCADE")
	require.NoError(t, err)

	// Run migrations
	err = database.Migrate(db)
	require.NoError(t, err)

	return db
}

func setupTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}

	targetingService := services.NewTargetingService(db)
	deliveryController := controllers.NewDeliveryController(targetingService)

	router := mux.NewRouter()
	router.HandleFunc("/v1/delivery", deliveryController.GetCampaigns).Methods("GET")

	server := httptest.NewServer(router)
	return server, db
}

func TestGetCampaigns_Success(t *testing.T) {
	server, db := setupTestServer(t)
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCIDs   []string
	}{
		{
			name:           "Germany Android request matches Duolingo",
			url:            "/v1/delivery?app=com.abc.xyz&country=germany&os=android",
			expectedStatus: http.StatusOK,
			expectedCIDs:   []string{"duolingo"},
		},
		{
			name:           "US Android Ludo King matches Spotify and Subway Surfer",
			url:            "/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android",
			expectedStatus: http.StatusOK,
			expectedCIDs:   []string{"spotify", "subwaysurfer"},
		},
		{
			name:           "US iOS request matches Spotify only",
			url:            "/v1/delivery?app=com.abc.xyz&country=us&os=ios",
			expectedStatus: http.StatusOK,
			expectedCIDs:   []string{"spotify"},
		},
		{
			name:           "India Android request matches Duolingo only",
			url:            "/v1/delivery?app=com.abc.xyz&country=india&os=android",
			expectedStatus: http.StatusOK,
			expectedCIDs:   []string{"duolingo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var campaigns []models.CampaignResponse
				err = json.NewDecoder(resp.Body).Decode(&campaigns)
				require.NoError(t, err)

				assert.Len(t, campaigns, len(tt.expectedCIDs))

				// Check if all expected campaign IDs are present
				actualCIDs := make([]string, len(campaigns))
				for i, campaign := range campaigns {
					actualCIDs[i] = campaign.CID
				}

				for _, expectedCID := range tt.expectedCIDs {
					assert.Contains(t, actualCIDs, expectedCID)
				}
			}
		})
	}
}

func TestGetCampaigns_NoMatch(t *testing.T) {
	server, db := setupTestServer(t)
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	// Request that shouldn't match any campaigns
	resp, err := http.Get(server.URL + "/v1/delivery?app=com.unknown.app&country=antarctica&os=windows")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetCampaigns_MissingParameters(t *testing.T) {
	server, db := setupTestServer(t)
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing app parameter",
			url:            "/v1/delivery?country=us&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing app param",
		},
		{
			name:           "Missing country parameter",
			url:            "/v1/delivery?app=com.test.app&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing country param",
		},
		{
			name:           "Missing os parameter",
			url:            "/v1/delivery?app=com.test.app&country=us",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing os param",
		},
		{
			name:           "Missing all parameters",
			url:            "/v1/delivery",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing app param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.url)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var errorResp models.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedError, errorResp.Error)
		})
	}
}

func TestTargetingLogic(t *testing.T) {
	server, db := setupTestServer(t)
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	// Test case-insensitive matching
	resp, err := http.Get(server.URL + "/v1/delivery?app=com.abc.xyz&country=GERMANY&os=ANDROID")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var campaigns []models.CampaignResponse
	err = json.NewDecoder(resp.Body).Decode(&campaigns)
	require.NoError(t, err)

	assert.Len(t, campaigns, 1)
	assert.Equal(t, "duolingo", campaigns[0].CID)
}

func TestInactiveCampaigns(t *testing.T) {
	server, db := setupTestServer(t)
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	// First, verify the campaign is initially active and matches
	resp, err := http.Get(server.URL + "/v1/delivery?app=com.abc.xyz&country=us&os=ios")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should initially return campaigns (assuming Spotify matches)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Now deactivate Spotify campaign
	result, err := db.Exec("UPDATE campaigns SET status = 'INACTIVE' WHERE id = 'spotify'")
	require.NoError(t, err)

	// Verify the update actually happened
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Greater(t, rowsAffected, int64(0), "No rows were updated - campaign 'spotify' might not exist")

	// Verify the status change in database
	var status string
	err = db.QueryRow("SELECT status FROM campaigns WHERE id = 'spotify'").Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "INACTIVE", status)

	// Force cache refresh instead of waiting
	// If your TargetingService has a method to force refresh, use it
	// Otherwise, wait longer for the cache to refresh
	time.Sleep(35 * time.Second) // Wait longer than cache refresh rate (30s)

	// Request that would normally match Spotify
	resp2, err := http.Get(server.URL + "/v1/delivery?app=com.abc.xyz&country=us&os=ios")
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Debug: Print response body if test fails
	if resp2.StatusCode != http.StatusNoContent {
		var campaigns []models.CampaignResponse
		json.NewDecoder(resp2.Body).Decode(&campaigns)
		t.Logf("Unexpected campaigns returned: %+v", campaigns)

		// Also check database state
		var dbCampaigns []string
		rows, _ := db.Query("SELECT id, status FROM campaigns")
		for rows.Next() {
			var id, status string
			rows.Scan(&id, &status)
			dbCampaigns = append(dbCampaigns, fmt.Sprintf("%s:%s", id, status))
		}
		rows.Close()
		t.Logf("Database campaigns: %v", dbCampaigns)
	}

	// Should return no content since Spotify is inactive
	assert.Equal(t, http.StatusNoContent, resp2.StatusCode)
}

// Benchmark tests for performance
func BenchmarkGetCampaigns(b *testing.B) {
	t := &testing.T{}
	server, db := setupTestServer(t)
	if server == nil || db == nil {
		b.Skip("Database not available for benchmark")
	}
	defer func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			db.Close()
		}
	}()

	url := server.URL + "/v1/delivery?app=com.abc.xyz&country=germany&os=android"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(url)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkTargetingService(b *testing.B) {
	t := &testing.T{}
	db := setupTestDB(t)
	if db == nil {
		b.Skip("Database not available for benchmark")
	}
	defer db.Close()

	targetingService := services.NewTargetingService(db)
	req := &models.DeliveryRequest{
		App:     "com.abc.xyz",
		Country: "germany",
		OS:      "android",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := targetingService.GetMatchingCampaigns(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
