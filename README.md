Targeting Engine
A high-performance microservice for campaign targeting and delivery, built with Go. This service routes the right campaigns to the right requests based on configurable targeting rules.
Features

High Performance: In-memory caching with sub-millisecond response times
Scalable Architecture: Designed for billions of requests and thousands of campaigns
Real-time Updates: Campaigns and targeting rules are updated in real-time
Comprehensive Monitoring: Prometheus metrics and Grafana dashboards
Robust Testing: Unit tests, integration tests, and benchmarks
Production Ready: Docker containerization and health checks

Architecture
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Application   │    │    Database     │
│                 │───▶│    (Go API)     │◄──▶│   PostgreSQL    │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Redis Cache   │
                       │   (Optional)    │
                       └─────────────────┘
Project Structure
targeting-engine/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── controllers/
│   │   └── delivery_controller.go  # HTTP handlers
│   ├── services/
│   │   └── targeting_service.go    # Business logic
│   ├── models/
│   │   └── models.go              # Data structures
│   ├── database/
│   │   ├── connection.go          # Database connection
│   │   └── migrations.go          # Database migrations
│   ├── cache/
│   │   └── redis_cache.go         # Redis caching
│   ├── monitoring/
│   │   └── metrics.go             # Prometheus metrics
│   └── utils/
│       └── response.go            # Helper functions
├── pkg/
│   └── config/
│       └── config.go              # Configuration management
├── tests/
│   └── delivery_test.go           # Test cases
├── docker-compose.yml             # Docker composition
├── Dockerfile                     # Container definition
└── README.md                      # This file
Quick Start
Using Docker Compose (Recommended)

Clone the repository

bashgit clone <repository-url>
cd targeting-engine

Start all services

bashdocker-compose up -d

Test the API

bashcurl "http://localhost:8080/v1/delivery?app=com.abc.xyz&country=germany&os=android"
Manual Setup

Prerequisites

Go 1.21+
PostgreSQL 12+
Redis 6+ (optional)


Database Setup

bashcreatedb targeting_engine

Environment Variables

bashexport DATABASE_URL="postgres://user:password@localhost/targeting_engine?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export PORT="8080"

Run the Application

bashgo mod download
go run cmd/server/main.go
API Documentation
Delivery Endpoint
GET /v1/delivery
Retrieves campaigns that match the targeting criteria.
Parameters
ParameterTypeRequiredDescriptionappstringYesApplication ID (e.g., "com.dream11.fantasy")countrystringYesUser's country (e.g., "US", "India")osstringYesOperating system (e.g., "Android", "iOS")
Response
Success (200)
json[
  {
    "cid": "spotify",
    "img": "https://somelink",
    "cta": "Download"
  }
]
No Content (204)
No campaigns match the criteria
Bad Request (400)
json{
  "error": "missing app param"
}
Examples
bash# Request matching multiple campaigns
curl "http://localhost:8080/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android"

# Request with no matches
curl "http://localhost:8080/v1/delivery?app=com.unknown.app&country=antarctica&os=windows"

# Invalid request
curl "http://localhost:8080/v1/delivery?country=us&os=android"
Database Schema
Campaigns Table
sqlCREATE TABLE campaigns (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    cta VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
Targeting Rules Table
sqlCREATE TABLE targeting_rules (
    id SERIAL PRIMARY KEY,
    campaign_id VARCHAR(255) NOT NULL,
    dimension VARCHAR(50) NOT NULL,    -- 'country', 'os', 'app'
    rule_type VARCHAR(50) NOT NULL,    -- 'include', 'exclude'
    values JSONB NOT NULL,            -- ["US", "Canada"]
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
);
Targeting Rules
Rule Types

Include: Campaign matches if request value is in the list
Exclude: Campaign matches if request value is NOT in the list

Dimensions

country: User's country code
os: Operating system (Android, iOS, Web)
app: Application identifier

Examples
sql-- Spotify: Include US and Canada
INSERT INTO targeting_rules (campaign_id, dimension, rule_type, values) 
VALUES ('spotify', 'country', 'include', '["US", "Canada"]');

-- Duolingo: Include Android/iOS but exclude US
INSERT INTO targeting_rules (campaign_id, dimension, rule_type, values) 
VALUES ('duolingo', 'os', 'include', '["Android", "iOS"]');
INSERT INTO targeting_rules (campaign_id, dimension, rule_type, values) 
VALUES ('duolingo', 'country', 'exclude', '["US"]');
Performance Optimizations
In-Memory Caching

All campaigns and targeting rules are cached in memory
Cache refreshes every 30 seconds automatically
Sub-millisecond response times for cached data

Database Optimizations

Indexed columns for fast queries
Connection pooling with configurable limits
Optimized schema design

Scalability Features

Stateless design for horizontal scaling
Efficient memory usage
Minimal database queries during serving

Monitoring
Metrics Endpoint
GET /metrics
Available Metrics

http_request_duration_seconds: Request latency histogram
http_requests_total: Total request count
active_campaigns_total: Number of active campaigns
cache_hit_ratio: Cache efficiency percentage
database_connections: Database connection pool stats

Grafana Dashboard
Access Grafana at http://localhost:3000 (admin/admin)
Testing
Run Tests
bashgo test ./tests/...
Run Benchmarks
bashgo test -bench=. ./tests/...
Test Coverage
bashgo test -cover ./...
Configuration
Environment Variables
VariableDefaultDescriptionDATABASE_URLpostgres://...PostgreSQL connection stringREDIS_URLredis://localhost:6379Redis connection stringPORT8080HTTP server portENVIRONMENTdevelopmentEnvironment (development/production)
Deployment
Docker
bashdocker build -t targeting-engine .
docker run -p 8080:8080 targeting-engine
Kubernetes
yamlapiVersion: apps/v1
kind: Deployment
metadata:
  name: targeting-engine
spec:
  replicas: 3
  selector:
    matchLabels:
      app: targeting-engine
  template:
    metadata:
      labels:
        app: targeting-engine
    spec:
      containers:
      - name: targeting-engine
        image: targeting-engine:latest
        ports:
        - containerPort: 8080
Development
Adding New Dimensions

Update the models.go constants
Modify the matchesDimensionRules function
Update database migrations
Add test cases

Adding New Features

Follow the existing project structure
Add appropriate tests
Update documentation

Contributing

Fork the repository
Create a feature branch
Make your changes
Add tests
Submit a pull request