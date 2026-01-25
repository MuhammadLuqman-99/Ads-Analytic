package postgres

import (
	"context"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// MetricsRepository implements repository.MetricsRepository
type MetricsRepository struct {
	db *Database
}

func NewMetricsRepository(db *Database) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// Campaign metrics
func (r *MetricsRepository) CreateCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *MetricsRepository) GetCampaignMetrics(ctx context.Context, campaignID uuid.UUID, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	var metrics []entity.CampaignMetricsDaily
	if err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND metric_date BETWEEN ? AND ?", campaignID, dateRange.StartDate, dateRange.EndDate).
		Order("metric_date ASC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *MetricsRepository) UpsertCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "campaign_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).Create(metrics).Error
}

func (r *MetricsRepository) BulkUpsertCampaignMetrics(ctx context.Context, metrics []entity.CampaignMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "campaign_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).CreateInBatches(metrics, 100).Error
}

// Ad set metrics
func (r *MetricsRepository) CreateAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *MetricsRepository) GetAdSetMetrics(ctx context.Context, adSetID uuid.UUID, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	var metrics []entity.AdSetMetricsDaily
	if err := r.db.WithContext(ctx).
		Where("ad_set_id = ? AND metric_date BETWEEN ? AND ?", adSetID, dateRange.StartDate, dateRange.EndDate).
		Order("metric_date ASC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *MetricsRepository) UpsertAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_set_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).Create(metrics).Error
}

func (r *MetricsRepository) BulkUpsertAdSetMetrics(ctx context.Context, metrics []entity.AdSetMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_set_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).CreateInBatches(metrics, 100).Error
}

// Ad metrics
func (r *MetricsRepository) CreateAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return r.db.WithContext(ctx).Create(metrics).Error
}

func (r *MetricsRepository) GetAdMetrics(ctx context.Context, adID uuid.UUID, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	var metrics []entity.AdMetricsDaily
	if err := r.db.WithContext(ctx).
		Where("ad_id = ? AND metric_date BETWEEN ? AND ?", adID, dateRange.StartDate, dateRange.EndDate).
		Order("metric_date ASC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *MetricsRepository) UpsertAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).Create(metrics).Error
}

func (r *MetricsRepository) BulkUpsertAdMetrics(ctx context.Context, metrics []entity.AdMetricsDaily) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ad_id"}, {Name: "metric_date"}},
		UpdateAll: true,
	}).CreateInBatches(metrics, 100).Error
}

// Aggregated metrics
func (r *MetricsRepository) GetAggregatedMetrics(ctx context.Context, filter entity.MetricsFilter) (*entity.AggregatedMetrics, error) {
	type result struct {
		TotalSpend       float64
		TotalImpressions int64
		TotalClicks      int64
		TotalConversions int64
		TotalRevenue     float64
	}
	var res result

	query := r.db.WithContext(ctx).Model(&entity.CampaignMetricsDaily{}).
		Select(`
			COALESCE(SUM(spend), 0) as total_spend,
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks,
			COALESCE(SUM(conversions), 0) as total_conversions,
			COALESCE(SUM(conversion_value), 0) as total_revenue
		`)

	if filter.OrganizationID != uuid.Nil {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if len(filter.Platforms) > 0 {
		query = query.Where("platform IN ?", filter.Platforms)
	}
	query = query.Where("metric_date BETWEEN ? AND ?", filter.DateRange.StartDate, filter.DateRange.EndDate)

	if err := query.Scan(&res).Error; err != nil {
		return nil, err
	}

	metrics := &entity.AggregatedMetrics{
		TotalImpressions: res.TotalImpressions,
		TotalClicks:      res.TotalClicks,
		TotalConversions: res.TotalConversions,
	}

	// Calculate derived metrics
	if res.TotalImpressions > 0 {
		metrics.AverageCTR = float64(res.TotalClicks) / float64(res.TotalImpressions) * 100
	}
	if res.TotalSpend > 0 {
		metrics.OverallROAS = res.TotalRevenue / res.TotalSpend
	}

	return metrics, nil
}

func (r *MetricsRepository) GetMetricsByPlatform(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) ([]entity.PlatformMetricsSummary, error) {
	var summaries []entity.PlatformMetricsSummary

	if err := r.db.WithContext(ctx).Model(&entity.CampaignMetricsDaily{}).
		Select(`
			platform,
			COALESCE(SUM(spend), 0) as spend,
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(conversions), 0) as conversions
		`).
		Where("organization_id = ? AND metric_date BETWEEN ? AND ?", orgID, dateRange.StartDate, dateRange.EndDate).
		Group("platform").
		Scan(&summaries).Error; err != nil {
		return nil, err
	}

	// Calculate derived metrics for each platform
	for i := range summaries {
		if summaries[i].Impressions > 0 {
			summaries[i].CTR = float64(summaries[i].Clicks) / float64(summaries[i].Impressions) * 100
		}
	}

	return summaries, nil
}

func (r *MetricsRepository) GetDailyTrend(ctx context.Context, filter entity.MetricsFilter) ([]entity.DailyMetricsTrend, error) {
	var trends []entity.DailyMetricsTrend

	query := r.db.WithContext(ctx).Model(&entity.CampaignMetricsDaily{}).
		Select(`
			metric_date as date,
			COALESCE(SUM(spend), 0) as spend,
			COALESCE(SUM(impressions), 0) as impressions,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(conversions), 0) as conversions
		`)

	if filter.OrganizationID != uuid.Nil {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if len(filter.Platforms) > 0 {
		query = query.Where("platform IN ?", filter.Platforms)
	}

	if err := query.
		Where("metric_date BETWEEN ? AND ?", filter.DateRange.StartDate, filter.DateRange.EndDate).
		Group("metric_date").
		Order("metric_date ASC").
		Scan(&trends).Error; err != nil {
		return nil, err
	}

	return trends, nil
}

func (r *MetricsRepository) GetTopPerformingCampaigns(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange, limit int) ([]entity.TopPerformer, error) {
	var performers []entity.TopPerformer

	if err := r.db.WithContext(ctx).Model(&entity.CampaignMetricsDaily{}).
		Select(`
			campaign_id as id,
			platform,
			COALESCE(SUM(spend), 0) as spend,
			COALESCE(SUM(conversions), 0) as conversions,
			CASE WHEN SUM(spend) > 0 THEN SUM(conversion_value) / SUM(spend) ELSE 0 END as roas
		`).
		Where("organization_id = ? AND metric_date BETWEEN ? AND ?", orgID, dateRange.StartDate, dateRange.EndDate).
		Group("campaign_id, platform").
		Order("roas DESC").
		Limit(limit).
		Scan(&performers).Error; err != nil {
		return nil, err
	}

	return performers, nil
}

var _ repository.MetricsRepository = (*MetricsRepository)(nil)

// AdAccountRepository implements repository.AdAccountRepository
type AdAccountRepository struct {
	db *Database
}

func NewAdAccountRepository(db *Database) *AdAccountRepository {
	return &AdAccountRepository{db: db}
}

func (r *AdAccountRepository) Create(ctx context.Context, account *entity.AdAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AdAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AdAccount, error) {
	var account entity.AdAccount
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AdAccountRepository) GetByPlatformID(ctx context.Context, connectedAccountID uuid.UUID, platformAdAccountID string) (*entity.AdAccount, error) {
	var account entity.AdAccount
	if err := r.db.WithContext(ctx).First(&account,
		"connected_account_id = ? AND platform_ad_account_id = ?",
		connectedAccountID, platformAdAccountID).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AdAccountRepository) Update(ctx context.Context, account *entity.AdAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *AdAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.AdAccount{}, "id = ?", id).Error
}

func (r *AdAccountRepository) ListByConnectedAccount(ctx context.Context, connectedAccountID uuid.UUID) ([]entity.AdAccount, error) {
	var accounts []entity.AdAccount
	if err := r.db.WithContext(ctx).Where("connected_account_id = ?", connectedAccountID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *AdAccountRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.AdAccount, error) {
	var accounts []entity.AdAccount
	if err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *AdAccountRepository) Upsert(ctx context.Context, account *entity.AdAccount) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "connected_account_id"}, {Name: "platform_ad_account_id"}},
		UpdateAll: true,
	}).Create(account).Error
}

var _ repository.AdAccountRepository = (*AdAccountRepository)(nil)
