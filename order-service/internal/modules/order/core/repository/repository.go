package repository

import (
	"context"
	"encoding/json"
	"log"

	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/entity"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewRepository(db *gorm.DB, cache *redis.Client) *Repository {
	db.AutoMigrate(&entity.Order{}, &entity.OrderRoute{})
	return &Repository{
		db:    db,
		cache: cache,
	}
}

func (r *Repository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *Repository) FindByID(ctx context.Context, id uint64) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", id).Update("status", status).Error
}

// PublishEvent mempublikasikan event order baru ke Redis Streams
func (r *Repository) PublishOrderCreatedEvent(ctx context.Context, order *entity.Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// Menulis data ke stream 'order.events' dengan key 'order.created'
	err = r.cache.XAdd(ctx, &redis.XAddArgs{
		Stream: "order_events_stream",
		Values: map[string]interface{}{
			"event_type": "order.created",
			"payload":    string(orderJSON),
		},
	}).Err()

	if err != nil {
		log.Printf("Gagal mem-publish event ke Redis Streams: %v", err)
		return err
	}

	log.Printf("Sukses mem-publish event order.created untuk ID %d", order.ID)
	return nil
}
