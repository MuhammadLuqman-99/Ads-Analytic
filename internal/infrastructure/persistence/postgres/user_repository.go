package postgres

import (
	"context"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository implements repository.UserRepository
type UserRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *UserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("email_verified_at", gorm.Expr("NOW()")).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName, phone string) error {
	updates := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"phone":      phone,
	}
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(updates).Error
}

var _ repository.UserRepository = (*UserRepository)(nil)

// OrganizationRepository implements repository.OrganizationRepository
type OrganizationRepository struct {
	db *Database
}

func NewOrganizationRepository(db *Database) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *entity.Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	var org entity.Organization
	if err := r.db.WithContext(ctx).First(&org, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	var org entity.Organization
	if err := r.db.WithContext(ctx).First(&org, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Organization{}, "id = ?", id).Error
}

func (r *OrganizationRepository) List(ctx context.Context, pagination *entity.Pagination) ([]entity.Organization, error) {
	var orgs []entity.Organization
	query := r.db.WithContext(ctx).Model(&entity.Organization{})
	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.PageSize)
	}
	if err := query.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *OrganizationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Organization, error) {
	var orgs []entity.Organization
	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ? AND organization_members.is_active = ?", userID, true).
		Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

var _ repository.OrganizationRepository = (*OrganizationRepository)(nil)

// OrganizationMemberRepository implements repository.OrganizationMemberRepository
type OrganizationMemberRepository struct {
	db *Database
}

func NewOrganizationMemberRepository(db *Database) *OrganizationMemberRepository {
	return &OrganizationMemberRepository{db: db}
}

func (r *OrganizationMemberRepository) Create(ctx context.Context, member *entity.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *OrganizationMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.OrganizationMember, error) {
	var member entity.OrganizationMember
	if err := r.db.WithContext(ctx).First(&member, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *OrganizationMemberRepository) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*entity.OrganizationMember, error) {
	var member entity.OrganizationMember
	if err := r.db.WithContext(ctx).First(&member, "organization_id = ? AND user_id = ?", orgID, userID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *OrganizationMemberRepository) Update(ctx context.Context, member *entity.OrganizationMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *OrganizationMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.OrganizationMember{}, "id = ?", id).Error
}

func (r *OrganizationMemberRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.OrganizationMember, error) {
	var members []entity.OrganizationMember
	query := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Preload("User")
	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.PageSize)
	}
	if err := query.Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *OrganizationMemberRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.OrganizationMember, error) {
	var members []entity.OrganizationMember
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).
		Preload("Organization").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *OrganizationMemberRepository) UpdateRole(ctx context.Context, id uuid.UUID, role entity.UserRole) error {
	return r.db.WithContext(ctx).Model(&entity.OrganizationMember{}).
		Where("id = ?", id).Update("role", role).Error
}

var _ repository.OrganizationMemberRepository = (*OrganizationMemberRepository)(nil)
