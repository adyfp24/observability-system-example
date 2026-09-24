package repository

import (
	"context"

	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/entity"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewRepository(db *gorm.DB, cache *redis.Client) *Repository {
	// Auto migrate tabel di startup jika diperlukan
	db.AutoMigrate(&entity.User{}, &entity.Driver{})
	return &Repository{
		db:    db,
		cache: cache,
	}
}

func (r *Repository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) CreateDriver(ctx context.Context, driver *entity.Driver) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

func (r *Repository) FindByPhoneNumber(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("phone_number = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindDriverByUserID(ctx context.Context, userID uint64) (*entity.Driver, error) {
	var driver entity.Driver
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}
