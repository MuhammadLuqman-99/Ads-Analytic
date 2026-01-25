package postgres

import (
	"context"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// ConnectedAccountRepository implements repository.ConnectedAccountRepository
type ConnectedAccountRepository struct {
	db *Database
}

func NewConnectedAccountRepository(db *Database) *ConnectedAccountRepository {
	return &ConnectedAccountRepository{db: db}
}

func (r *ConnectedAccountRepository) Create(ctx context.Context, account *entity.ConnectedAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *ConnectedAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ConnectedAccount, error) {
	var account entity.ConnectedAccount
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *ConnectedAccountRepository) GetByPlatformAccountID(ctx context.Context, orgID uuid.UUID, platform entity.Platform, platformAccountID string) (*entity.ConnectedAccount, error) {
	var account entity.ConnectedAccount
	if err := r.db.WithContext(ctx).First(&account,
		"organization_id = ? AND platform = ? AND platform_account_id = ?",
		orgID, platform, platformAccountID).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *ConnectedAccountRepository) Update(ctx context.Context, account *entity.ConnectedAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *ConnectedAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.ConnectedAccount{}, "id = ?", id).Error
}

func (r *ConnectedAccountRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error) {
	var accounts []entity.ConnectedAccount
	if err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *ConnectedAccountRepository) ListByPlatform(ctx context.Context, orgID uuid.UUID, platform entity.Platform) ([]entity.ConnectedAccount, error) {
	var accounts []entity.ConnectedAccount
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND platform = ?", orgID, platform).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *ConnectedAccountRepository) ListExpiring(ctx context.Context, withinMinutes int) ([]entity.ConnectedAccount, error) {
	var accounts []entity.ConnectedAccount
	expiryThreshold := time.Now().Add(time.Duration(withinMinutes) * time.Minute)
	if err := r.db.WithContext(ctx).
		Where("status = ? AND token_expires_at IS NOT NULL AND token_expires_at <= ?",
			entity.AccountStatusActive, expiryThreshold).
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *ConnectedAccountRepository) ListActive(ctx context.Context) ([]entity.ConnectedAccount, error) {
	var accounts []entity.ConnectedAccount
	if err := r.db.WithContext(ctx).Where("status = ?", entity.AccountStatusActive).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *ConnectedAccountRepository) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.ConnectedAccount{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"access_token":     accessToken,
			"refresh_token":    refreshToken,
			"token_expires_at": expiresAt,
			"updated_at":       time.Now(),
		}).Error
}

func (r *ConnectedAccountRepository) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus, syncError string) error {
	return r.db.WithContext(ctx).Model(&entity.ConnectedAccount{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"sync_error": syncError,
			"updated_at": time.Now(),
		}).Error
}

func (r *ConnectedAccountRepository) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.ConnectedAccount{}).
		Where("id = ?", id).
		Update("last_synced_at", now).Error
}

var _ repository.ConnectedAccountRepository = (*ConnectedAccountRepository)(nil)

// CampaignRepository implements repository.CampaignRepository
type CampaignRepository struct {
	db *Database
}

func NewCampaignRepository(db *Database) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Create(campaign).Error
}

func (r *CampaignRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.WithContext(ctx).First(&campaign, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetByPlatformID(ctx context.Context, adAccountID uuid.UUID, platformCampaignID string) (*entity.Campaign, error) {
	var campaign entity.Campaign
	if err := r.db.WithContext(ctx).First(&campaign, "ad_account_id = ? AND platform_campaign_id = ?", adAccountID, platformCampaignID).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Save(campaign).Error
}

func (r *CampaignRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Campaign{}, "id = ?", id).Error
}

func (r *CampaignRepository) List(ctx context.Context, filter entity.CampaignFilter) ([]entity.Campaign, int64, error) {
	var campaigns []entity.Campaign
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Campaign{})

	if filter.OrganizationID != uuid.Nil {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if len(filter.Platforms) > 0 {
		query = query.Where("platform IN ?", filter.Platforms)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	if filter.SearchTerm != "" {
		query = query.Where("platform_campaign_name ILIKE ?", "%"+filter.SearchTerm+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Pagination != nil {
		query = query.Offset(filter.Pagination.Offset()).Limit(filter.Pagination.PageSize)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

func (r *CampaignRepository) ListByAdAccount(ctx context.Context, adAccountID uuid.UUID) ([]entity.Campaign, error) {
	var campaigns []entity.Campaign
	if err := r.db.WithContext(ctx).Where("ad_account_id = ?", adAccountID).Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *CampaignRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.Campaign, error) {
	var campaigns []entity.Campaign
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID)
	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.PageSize)
	}
	if err := query.Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *CampaignRepository) Upsert(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_account_id"}, {Name: "platform_campaign_id"}},
		UpdateAll: true,
	}).Create(campaign).Error
}

func (r *CampaignRepository) BulkUpsert(ctx context.Context, campaigns []entity.Campaign) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_account_id"}, {Name: "platform_campaign_id"}},
		UpdateAll: true,
	}).CreateInBatches(campaigns, 100).Error
}

func (r *CampaignRepository) GetSummaries(ctx context.Context, filter entity.CampaignFilter) ([]entity.CampaignSummary, error) {
	// This would use a complex query joining with metrics
	return nil, nil
}

func (r *CampaignRepository) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Campaign{}).
		Where("id = ?", id).
		Update("last_synced_at", time.Now()).Error
}

var _ repository.CampaignRepository = (*CampaignRepository)(nil)
