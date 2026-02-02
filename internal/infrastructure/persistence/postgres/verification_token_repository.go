package postgres

import (
	"context"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationTokenRepository implements repository.VerificationTokenRepository
type VerificationTokenRepository struct {
	db *Database
}

// NewVerificationTokenRepository creates a new verification token repository
func NewVerificationTokenRepository(db *Database) *VerificationTokenRepository {
	return &VerificationTokenRepository{db: db}
}

func (r *VerificationTokenRepository) Create(ctx context.Context, token *entity.VerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *VerificationTokenRepository) GetByToken(ctx context.Context, token string) (*entity.VerificationToken, error) {
	var verificationToken entity.VerificationToken
	if err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&verificationToken).Error; err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *VerificationTokenRepository) GetByUserAndType(ctx context.Context, userID uuid.UUID, tokenType entity.TokenType) (*entity.VerificationToken, error) {
	var verificationToken entity.VerificationToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ? AND is_used = ?", userID, tokenType, false).
		Order("created_at DESC").
		First(&verificationToken).Error; err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *VerificationTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.VerificationToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_used": true,
			"used_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *VerificationTokenRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.VerificationToken{}).Error
}

func (r *VerificationTokenRepository) DeleteByUserAndType(ctx context.Context, userID uuid.UUID, tokenType entity.TokenType) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ?", userID, tokenType).
		Delete(&entity.VerificationToken{}).Error
}

func (r *VerificationTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.VerificationToken{}).Error
}

var _ repository.VerificationTokenRepository = (*VerificationTokenRepository)(nil)
