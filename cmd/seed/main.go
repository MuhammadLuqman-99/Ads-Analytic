// =============================================================================
// Seed Data Script
// Populates the database with realistic fake data for testing
// =============================================================================

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// =============================================================================
// Models (simplified for seeding)
// =============================================================================

// JSONMap is a custom type for JSONB columns
type JSONMap map[string]interface{}

type Organization struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key"`
	Name                  string    `gorm:"size:255;not null"`
	Slug                  string    `gorm:"size:100;unique;not null"`
	LogoURL               string    `gorm:"size:500"`
	SubscriptionPlan      string    `gorm:"type:subscription_plan;default:'free'"`
	SubscriptionExpiresAt *time.Time
	Settings              JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	Metadata              JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	IsActive              bool    `gorm:"default:true"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	Email           string    `gorm:"size:255;unique;not null"`
	PasswordHash    string    `gorm:"size:255"`
	FirstName       string    `gorm:"size:100"`
	LastName        string    `gorm:"size:100"`
	AvatarURL       string    `gorm:"size:500"`
	Phone           string    `gorm:"size:20"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	IsActive        bool    `gorm:"default:true"`
	Metadata        JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrganizationMember struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	Role           string    `gorm:"type:user_role;default:'viewer'"`
	InvitedBy      *uuid.UUID
	InvitedAt      *time.Time
	JoinedAt       *time.Time
	IsActive       bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConnectedAccount struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key"`
	OrganizationID      uuid.UUID `gorm:"type:uuid;not null"`
	Platform            string    `gorm:"type:platform_type;not null"`
	PlatformAccountID   string    `gorm:"size:255;not null"`
	PlatformAccountName string    `gorm:"size:255"`
	PlatformUserID      string    `gorm:"size:255"`
	Status              string    `gorm:"type:account_status;default:'active'"`
	LastSyncedAt        *time.Time
	SyncError           string
	AccountTimezone     string  `gorm:"size:50"`
	AccountCurrency     string  `gorm:"size:3;default:'MYR'"`
	Metadata            JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	AccessToken         string  `gorm:"type:text;not null"`
	RefreshToken        string  `gorm:"type:text"`
	TokenType           string  `gorm:"size:50;default:'Bearer'"`
	TokenExpiresAt      *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type AdAccount struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key"`
	ConnectedAccountID    uuid.UUID `gorm:"type:uuid;not null"`
	OrganizationID        uuid.UUID `gorm:"type:uuid;not null"`
	Platform              string    `gorm:"type:platform_type;not null"`
	PlatformAdAccountID   string    `gorm:"size:255;not null"`
	PlatformAdAccountName string    `gorm:"size:255"`
	Currency              string    `gorm:"size:3;default:'MYR'"`
	Timezone              string    `gorm:"size:50"`
	IsActive              bool      `gorm:"default:true"`
	LastSyncedAt          *time.Time
	Metadata              JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type Campaign struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key"`
	AdAccountID          uuid.UUID `gorm:"type:uuid;not null"`
	OrganizationID       uuid.UUID `gorm:"type:uuid;not null"`
	Platform             string    `gorm:"type:platform_type;not null"`
	PlatformCampaignID   string    `gorm:"size:255;not null"`
	PlatformCampaignName string    `gorm:"size:500"`
	Objective            string    `gorm:"type:campaign_objective"`
	Status               string    `gorm:"type:campaign_status;default:'active'"`
	DailyBudget          *decimal.Decimal
	LifetimeBudget       *decimal.Decimal
	BudgetCurrency       string     `gorm:"size:3;default:'MYR'"`
	StartDate            *time.Time `gorm:"type:date"`
	EndDate              *time.Time `gorm:"type:date"`
	PlatformData         JSONMap    `gorm:"type:jsonb;default:'{}';serializer:json"`
	PlatformCreatedAt    *time.Time
	PlatformUpdatedAt    *time.Time
	LastSyncedAt         *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CampaignMetricsDaily struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key"`
	CampaignID        uuid.UUID `gorm:"type:uuid;not null"`
	OrganizationID    uuid.UUID `gorm:"type:uuid;not null"`
	Platform          string    `gorm:"type:platform_type;not null"`
	MetricDate        time.Time `gorm:"type:date;not null"`
	Impressions       int64     `gorm:"default:0"`
	Reach             int64     `gorm:"default:0"`
	Clicks            int64     `gorm:"default:0"`
	UniqueClicks      int64     `gorm:"default:0"`
	Spend             decimal.Decimal
	Currency          string `gorm:"size:3;default:'MYR'"`
	Likes             int64  `gorm:"default:0"`
	Comments          int64  `gorm:"default:0"`
	Shares            int64  `gorm:"default:0"`
	Saves             int64  `gorm:"default:0"`
	VideoViews        int64  `gorm:"default:0"`
	VideoViewsP25     int64  `gorm:"default:0"`
	VideoViewsP50     int64  `gorm:"default:0"`
	VideoViewsP75     int64  `gorm:"default:0"`
	VideoViewsP100    int64  `gorm:"default:0"`
	Conversions       int64  `gorm:"default:0"`
	ConversionValue   decimal.Decimal
	AddToCart         int64 `gorm:"default:0"`
	CheckoutInitiated int64 `gorm:"default:0"`
	Purchases         int64 `gorm:"default:0"`
	PurchaseValue     decimal.Decimal
	CTR               *float64
	CPC               *decimal.Decimal
	CPM               *decimal.Decimal
	CPA               *decimal.Decimal
	ROAS              *float64
	PlatformMetrics JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	LastSyncedAt    *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName overrides the default table name
func (CampaignMetricsDaily) TableName() string {
	return "campaign_metrics_daily"
}

type SyncHistory struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	OrganizationID     uuid.UUID `gorm:"type:uuid;not null"`
	ConnectedAccountID uuid.UUID `gorm:"type:uuid;not null"`
	Platform           string    `gorm:"type:platform_type;not null"`
	SyncType           string    `gorm:"size:50;not null"`
	Status             string    `gorm:"size:20;not null"`
	StartedAt          time.Time
	CompletedAt        *time.Time
	RecordsProcessed   int
	ErrorMessage       string  `gorm:"type:text"`
	Metadata           JSONMap `gorm:"type:jsonb;default:'{}';serializer:json"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TableName overrides the default table name
func (SyncHistory) TableName() string {
	return "sync_history"
}

// =============================================================================
// Main
// =============================================================================

func main() {
	log.Println("Starting seed data script...")

	// Get database connection
	db := connectDB()

	// Run migrations (if tables don't exist)
	log.Println("Ensuring tables exist...")

	// Clean existing test data
	log.Println("Cleaning existing test data...")
	cleanTestData(db)

	// Seed data
	log.Println("Seeding organizations...")
	orgs := seedOrganizations(db)

	log.Println("Seeding users...")
	users := seedUsers(db, orgs)

	log.Println("Seeding connected accounts...")
	connectedAccounts := seedConnectedAccounts(db, orgs)

	log.Println("Seeding ad accounts...")
	adAccounts := seedAdAccounts(db, connectedAccounts)

	log.Println("Seeding campaigns...")
	campaigns := seedCampaigns(db, adAccounts, orgs)

	log.Println("Seeding daily metrics (90 days)...")
	seedMetrics(db, campaigns, 90)

	log.Println("Seeding sync history...")
	seedSyncHistory(db, connectedAccounts)

	log.Println("")
	log.Println("========================================")
	log.Println("Seed completed successfully!")
	log.Println("========================================")
	log.Println("")
	log.Println("Test accounts:")
	log.Println("  admin@test.com / password123 (Business plan, all platforms)")
	log.Println("  pro@test.com / password123 (Pro plan, Meta + TikTok)")
	log.Println("  free@test.com / password123 (Free plan, Meta only)")
	log.Println("")
	log.Printf("Created: %d organizations, %d users, %d campaigns\n", len(orgs), len(users), len(campaigns))
}

func connectDB() *gorm.DB {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "localpassword123")
	dbname := getEnv("DB_NAME", "ads_local")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanTestData(db *gorm.DB) {
	// Delete in reverse dependency order
	db.Exec("DELETE FROM campaign_metrics_daily WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM sync_history WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM campaigns WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM ad_accounts WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM connected_accounts WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM organization_members WHERE organization_id IN (SELECT id FROM organizations WHERE slug LIKE 'test-%')")
	db.Exec("DELETE FROM users WHERE email LIKE '%@test.com'")
	db.Exec("DELETE FROM organizations WHERE slug LIKE 'test-%'")
}

// =============================================================================
// Seed Functions
// =============================================================================

func seedOrganizations(db *gorm.DB) []Organization {
	now := time.Now()
	expiresAt := now.AddDate(1, 0, 0) // 1 year from now

	orgs := []Organization{
		{
			ID:                    uuid.New(),
			Name:                  "Business Corp Sdn Bhd",
			Slug:                  "test-business-corp",
			SubscriptionPlan:      "enterprise",
			SubscriptionExpiresAt: &expiresAt,
			IsActive:              true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    uuid.New(),
			Name:                  "Pro Marketing Agency",
			Slug:                  "test-pro-agency",
			SubscriptionPlan:      "professional",
			SubscriptionExpiresAt: &expiresAt,
			IsActive:              true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		{
			ID:                    uuid.New(),
			Name:                  "Free Startup",
			Slug:                  "test-free-startup",
			SubscriptionPlan:      "free",
			SubscriptionExpiresAt: nil,
			IsActive:              true,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
	}

	for _, org := range orgs {
		if err := db.Create(&org).Error; err != nil {
			log.Fatalf("Failed to create organization: %v", err)
		}
	}

	return orgs
}

func seedUsers(db *gorm.DB, orgs []Organization) []User {
	now := time.Now()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []User{
		{
			ID:              uuid.New(),
			Email:           "admin@test.com",
			PasswordHash:    string(passwordHash),
			FirstName:       "Admin",
			LastName:        "User",
			EmailVerifiedAt: &now,
			LastLoginAt:     &now,
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New(),
			Email:           "pro@test.com",
			PasswordHash:    string(passwordHash),
			FirstName:       "Pro",
			LastName:        "User",
			EmailVerifiedAt: &now,
			LastLoginAt:     &now,
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New(),
			Email:           "free@test.com",
			PasswordHash:    string(passwordHash),
			FirstName:       "Free",
			LastName:        "User",
			EmailVerifiedAt: &now,
			LastLoginAt:     &now,
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New(),
			Email:           "analyst@test.com",
			PasswordHash:    string(passwordHash),
			FirstName:       "Analyst",
			LastName:        "User",
			EmailVerifiedAt: &now,
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New(),
			Email:           "viewer@test.com",
			PasswordHash:    string(passwordHash),
			FirstName:       "Viewer",
			LastName:        "User",
			EmailVerifiedAt: &now,
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create organization memberships
	memberships := []OrganizationMember{
		// Business Corp - admin@test.com is owner
		{ID: uuid.New(), OrganizationID: orgs[0].ID, UserID: users[0].ID, Role: "owner", JoinedAt: &now, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), OrganizationID: orgs[0].ID, UserID: users[3].ID, Role: "analyst", JoinedAt: &now, IsActive: true, CreatedAt: now, UpdatedAt: now},

		// Pro Agency - pro@test.com is owner
		{ID: uuid.New(), OrganizationID: orgs[1].ID, UserID: users[1].ID, Role: "owner", JoinedAt: &now, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), OrganizationID: orgs[1].ID, UserID: users[4].ID, Role: "viewer", JoinedAt: &now, IsActive: true, CreatedAt: now, UpdatedAt: now},

		// Free Startup - free@test.com is owner
		{ID: uuid.New(), OrganizationID: orgs[2].ID, UserID: users[2].ID, Role: "owner", JoinedAt: &now, IsActive: true, CreatedAt: now, UpdatedAt: now},
	}

	for _, m := range memberships {
		if err := db.Create(&m).Error; err != nil {
			log.Fatalf("Failed to create membership: %v", err)
		}
	}

	return users
}

func seedConnectedAccounts(db *gorm.DB, orgs []Organization) []ConnectedAccount {
	now := time.Now()
	tokenExpiry := now.AddDate(0, 1, 0) // 1 month from now

	accounts := []ConnectedAccount{
		// Business Corp - All platforms (Meta x2, TikTok x2, Shopee x1)
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[0].ID,
			Platform:            "meta",
			PlatformAccountID:   "act_123456789",
			PlatformAccountName: "Business Corp - Meta Ads 1",
			PlatformUserID:      "meta_user_001",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("meta"),
			RefreshToken:        generateMockToken("meta_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[0].ID,
			Platform:            "meta",
			PlatformAccountID:   "act_987654321",
			PlatformAccountName: "Business Corp - Meta Ads 2",
			PlatformUserID:      "meta_user_001",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("meta"),
			RefreshToken:        generateMockToken("meta_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[0].ID,
			Platform:            "tiktok",
			PlatformAccountID:   "7123456789012345678",
			PlatformAccountName: "Business Corp - TikTok Ads 1",
			PlatformUserID:      "tiktok_user_001",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("tiktok"),
			RefreshToken:        generateMockToken("tiktok_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[0].ID,
			Platform:            "tiktok",
			PlatformAccountID:   "7987654321098765432",
			PlatformAccountName: "Business Corp - TikTok Ads 2",
			PlatformUserID:      "tiktok_user_001",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("tiktok"),
			RefreshToken:        generateMockToken("tiktok_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[0].ID,
			Platform:            "shopee",
			PlatformAccountID:   "shop_12345678",
			PlatformAccountName: "Business Corp Shopee Store",
			PlatformUserID:      "shopee_user_001",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("shopee"),
			RefreshToken:        generateMockToken("shopee_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},

		// Pro Agency - Meta + TikTok
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[1].ID,
			Platform:            "meta",
			PlatformAccountID:   "act_pro_meta_001",
			PlatformAccountName: "Pro Agency - Meta Ads",
			PlatformUserID:      "meta_user_002",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("meta"),
			RefreshToken:        generateMockToken("meta_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[1].ID,
			Platform:            "tiktok",
			PlatformAccountID:   "7111222333444555666",
			PlatformAccountName: "Pro Agency - TikTok Ads",
			PlatformUserID:      "tiktok_user_002",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("tiktok"),
			RefreshToken:        generateMockToken("tiktok_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},

		// Free Startup - Meta only
		{
			ID:                  uuid.New(),
			OrganizationID:      orgs[2].ID,
			Platform:            "meta",
			PlatformAccountID:   "act_free_meta_001",
			PlatformAccountName: "Free Startup - Meta Ads",
			PlatformUserID:      "meta_user_003",
			Status:              "active",
			LastSyncedAt:        &now,
			AccountTimezone:     "Asia/Kuala_Lumpur",
			AccountCurrency:     "MYR",
			AccessToken:         generateMockToken("meta"),
			RefreshToken:        generateMockToken("meta_refresh"),
			TokenType:           "Bearer",
			TokenExpiresAt:      &tokenExpiry,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}

	for _, acc := range accounts {
		if err := db.Create(&acc).Error; err != nil {
			log.Fatalf("Failed to create connected account: %v", err)
		}
	}

	return accounts
}

func seedAdAccounts(db *gorm.DB, connectedAccounts []ConnectedAccount) []AdAccount {
	now := time.Now()
	adAccounts := make([]AdAccount, 0, len(connectedAccounts))

	for _, ca := range connectedAccounts {
		adAccount := AdAccount{
			ID:                    uuid.New(),
			ConnectedAccountID:    ca.ID,
			OrganizationID:        ca.OrganizationID,
			Platform:              ca.Platform,
			PlatformAdAccountID:   ca.PlatformAccountID,
			PlatformAdAccountName: ca.PlatformAccountName,
			Currency:              "MYR",
			Timezone:              "Asia/Kuala_Lumpur",
			IsActive:              true,
			LastSyncedAt:          &now,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		if err := db.Create(&adAccount).Error; err != nil {
			log.Fatalf("Failed to create ad account: %v", err)
		}
		adAccounts = append(adAccounts, adAccount)
	}

	return adAccounts
}

func seedCampaigns(db *gorm.DB, adAccounts []AdAccount, orgs []Organization) []Campaign {
	now := time.Now()
	campaigns := make([]Campaign, 0, 50)

	objectives := []string{"conversions", "traffic", "engagement", "awareness", "leads", "sales", "video_views"}
	statuses := []string{"active", "active", "active", "paused", "paused"} // More active than paused

	campaignNames := []string{
		"Brand Awareness Campaign", "Product Launch", "Holiday Sale",
		"Retargeting - Cart Abandoners", "Lookalike Audience", "Video Views Campaign",
		"Lead Generation", "App Install Campaign", "Traffic Campaign",
		"Conversion Campaign", "Engagement Boost", "Seasonal Promotion",
		"New Customer Acquisition", "Re-engagement Campaign", "Flash Sale",
	}

	campaignCount := 0
	for _, adAccount := range adAccounts {
		// Determine number of campaigns per account
		numCampaigns := 5 + mathrand.Intn(8) // 5-12 campaigns per account

		for i := 0; i < numCampaigns && campaignCount < 50; i++ {
			objective := objectives[mathrand.Intn(len(objectives))]
			status := statuses[mathrand.Intn(len(statuses))]
			name := campaignNames[mathrand.Intn(len(campaignNames))]

			// Random start date (1-90 days ago)
			daysAgo := mathrand.Intn(90) + 1
			startDate := now.AddDate(0, 0, -daysAgo)

			// Some campaigns have ended
			var endDate *time.Time
			if status == "paused" || mathrand.Float32() < 0.2 {
				ed := startDate.AddDate(0, 0, mathrand.Intn(30)+7)
				if ed.Before(now) {
					endDate = &ed
					status = "paused"
				}
			}

			// Budget between RM50 - RM500/day
			dailyBudget := decimal.NewFromFloat(float64(mathrand.Intn(450)+50) + mathrand.Float64())

			campaign := Campaign{
				ID:                   uuid.New(),
				AdAccountID:          adAccount.ID,
				OrganizationID:       adAccount.OrganizationID,
				Platform:             adAccount.Platform,
				PlatformCampaignID:   fmt.Sprintf("%s_camp_%d", adAccount.PlatformAdAccountID, 1000+i),
				PlatformCampaignName: fmt.Sprintf("[%s] %s %d", adAccount.Platform, name, i+1),
				Objective:            objective,
				Status:               status,
				DailyBudget:          &dailyBudget,
				BudgetCurrency:       "MYR",
				StartDate:            &startDate,
				EndDate:              endDate,
				PlatformCreatedAt:    &startDate,
				PlatformUpdatedAt:    &now,
				LastSyncedAt:         &now,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			if err := db.Create(&campaign).Error; err != nil {
				log.Fatalf("Failed to create campaign: %v", err)
			}
			campaigns = append(campaigns, campaign)
			campaignCount++
		}
	}

	return campaigns
}

func seedMetrics(db *gorm.DB, campaigns []Campaign, days int) {
	now := time.Now()
	baseDate := now.AddDate(0, 0, -days)

	// Define campaign performance profiles
	type PerformanceProfile struct {
		SpendMultiplier    float64
		CTRBase            float64
		ConversionRateBase float64
		ROASTarget         float64
	}

	profiles := []PerformanceProfile{
		{SpendMultiplier: 1.0, CTRBase: 3.5, ConversionRateBase: 5.0, ROASTarget: 4.0}, // High performer
		{SpendMultiplier: 0.8, CTRBase: 2.5, ConversionRateBase: 3.5, ROASTarget: 2.5}, // Good performer
		{SpendMultiplier: 0.6, CTRBase: 1.8, ConversionRateBase: 2.0, ROASTarget: 1.5}, // Average
		{SpendMultiplier: 0.4, CTRBase: 1.2, ConversionRateBase: 1.0, ROASTarget: 0.8}, // Poor performer
	}

	totalMetrics := 0

	for _, campaign := range campaigns {
		// Assign a performance profile to this campaign
		profile := profiles[mathrand.Intn(len(profiles))]

		// Get campaign start date
		startDate := baseDate
		if campaign.StartDate != nil && campaign.StartDate.After(baseDate) {
			startDate = *campaign.StartDate
		}

		for d := 0; d < days; d++ {
			date := baseDate.AddDate(0, 0, d)

			// Skip if before campaign start
			if date.Before(startDate) {
				continue
			}

			// Skip if campaign has ended
			if campaign.EndDate != nil && date.After(*campaign.EndDate) {
				continue
			}

			// Randomly skip some days (5% chance - data gaps)
			if mathrand.Float32() < 0.05 {
				continue
			}

			// Skip if campaign is paused (50% chance of having data on paused days)
			if campaign.Status == "paused" && mathrand.Float32() < 0.5 {
				continue
			}

			// Generate metrics with some randomness around the profile
			spend := (float64(mathrand.Intn(400)+50) + mathrand.Float64()) * profile.SpendMultiplier

			// Impressions based on spend (roughly RM5-15 CPM)
			cpm := 5 + mathrand.Float64()*10
			impressions := int64(spend / cpm * 1000)

			// Add day-of-week variance (weekends slightly lower)
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				impressions = int64(float64(impressions) * 0.85)
			}

			// CTR with variance
			ctr := profile.CTRBase + (mathrand.Float64()-0.5)*1.0
			if ctr < 0.5 {
				ctr = 0.5
			}
			clicks := int64(float64(impressions) * ctr / 100)

			// Conversion rate with variance
			convRate := profile.ConversionRateBase + (mathrand.Float64()-0.5)*1.5
			if convRate < 0.2 {
				convRate = 0.2
			}
			conversions := int64(float64(clicks) * convRate / 100)

			// Revenue based on target ROAS
			avgOrderValue := 80 + mathrand.Float64()*70 // RM80-150
			purchaseValue := float64(conversions) * avgOrderValue

			// Adjust to hit roughly the target ROAS
			roasVariance := (mathrand.Float64() - 0.5) * 0.5
			targetRevenue := spend * (profile.ROASTarget + roasVariance)
			if targetRevenue > 0 {
				purchaseValue = targetRevenue
			}

			// Engagement metrics
			likes := int64(float64(impressions) * 0.002 * (1 + mathrand.Float64()))
			comments := int64(float64(likes) * 0.1)
			shares := int64(float64(likes) * 0.05)
			videoViews := int64(float64(impressions) * 0.3)

			// Calculate derived metrics
			ctrCalc := float64(clicks) / float64(impressions) * 100
			var cpcCalc, cpmCalc, cpaCalc *decimal.Decimal
			var roasCalc *float64

			if clicks > 0 {
				cpc := decimal.NewFromFloat(spend / float64(clicks))
				cpcCalc = &cpc
			}
			if impressions > 0 {
				cpmVal := decimal.NewFromFloat(spend / float64(impressions) * 1000)
				cpmCalc = &cpmVal
			}
			if conversions > 0 {
				cpa := decimal.NewFromFloat(spend / float64(conversions))
				cpaCalc = &cpa
			}
			if spend > 0 {
				roas := purchaseValue / spend
				roasCalc = &roas
			}

			metric := CampaignMetricsDaily{
				ID:              uuid.New(),
				CampaignID:      campaign.ID,
				OrganizationID:  campaign.OrganizationID,
				Platform:        campaign.Platform,
				MetricDate:      date,
				Impressions:     impressions,
				Reach:           int64(float64(impressions) * 0.7),
				Clicks:          clicks,
				UniqueClicks:    int64(float64(clicks) * 0.85),
				Spend:           decimal.NewFromFloat(spend),
				Currency:        "MYR",
				Likes:           likes,
				Comments:        comments,
				Shares:          shares,
				Saves:           int64(float64(likes) * 0.02),
				VideoViews:      videoViews,
				VideoViewsP25:   int64(float64(videoViews) * 0.6),
				VideoViewsP50:   int64(float64(videoViews) * 0.4),
				VideoViewsP75:   int64(float64(videoViews) * 0.25),
				VideoViewsP100:  int64(float64(videoViews) * 0.15),
				Conversions:     conversions,
				ConversionValue: decimal.NewFromFloat(purchaseValue),
				AddToCart:       conversions * 3,
				CheckoutInitiated: int64(math.Max(float64(conversions*2), 1)),
				Purchases:       conversions,
				PurchaseValue:   decimal.NewFromFloat(purchaseValue),
				CTR:             &ctrCalc,
				CPC:             cpcCalc,
				CPM:             cpmCalc,
				CPA:             cpaCalc,
				ROAS:            roasCalc,
				LastSyncedAt:    &now,
				CreatedAt:       now,
				UpdatedAt:       now,
			}

			if err := db.Create(&metric).Error; err != nil {
				log.Fatalf("Failed to create metric: %v", err)
			}
			totalMetrics++
		}
	}

	log.Printf("Created %d metric records", totalMetrics)
}

func seedSyncHistory(db *gorm.DB, connectedAccounts []ConnectedAccount) {
	now := time.Now()

	for _, ca := range connectedAccounts {
		// Create a few sync history records per account
		for i := 0; i < 5; i++ {
			startedAt := now.AddDate(0, 0, -i).Add(-time.Hour * time.Duration(mathrand.Intn(24)))
			completedAt := startedAt.Add(time.Duration(mathrand.Intn(300)+30) * time.Second)

			status := "completed"
			errorMessage := ""
			if mathrand.Float32() < 0.1 {
				status = "failed"
				errorMessage = "Rate limit exceeded, will retry"
			}

			syncHistory := SyncHistory{
				ID:                 uuid.New(),
				OrganizationID:     ca.OrganizationID,
				ConnectedAccountID: ca.ID,
				Platform:           ca.Platform,
				SyncType:           "full",
				Status:             status,
				StartedAt:          startedAt,
				CompletedAt:        &completedAt,
				RecordsProcessed:   mathrand.Intn(1000) + 100,
				ErrorMessage:       errorMessage,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := db.Create(&syncHistory).Error; err != nil {
				// Skip if table doesn't exist
				log.Printf("Warning: Could not create sync history (table may not exist): %v", err)
				return
			}
		}
	}
}

func generateMockToken(prefix string) string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("mock_%s_%s", prefix, hex.EncodeToString(bytes))
}
