package repository

import (
	"context"
	"errors"

	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/entity"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewRepository(db *gorm.DB, cache *redis.Client) *Repository {
	db.AutoMigrate(&entity.Wallet{}, &entity.Transaction{})
	return &Repository{
		db:    db,
		cache: cache,
	}
}

func (r *Repository) FindByUserID(ctx context.Context, userID uint64) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Auto create wallet jika belum ada (lazy initialization)
			wallet = entity.Wallet{UserID: userID, Balance: 0.0}
			if err := r.db.WithContext(ctx).Create(&wallet).Error; err != nil {
				return nil, err
			}
			return &wallet, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *Repository) UpdateBalance(ctx context.Context, userID uint64, amount float64, txType string, desc string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	// Menjalankan di dalam Transaksi Database GORM
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet entity.Wallet
		// Gunakan select for update untuk mengunci baris (prevent race condition)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}

		if txType == "debit" {
			if wallet.Balance < amount {
				return errors.New("saldo tidak mencukupi")
			}
			wallet.Balance -= amount
		} else {
			wallet.Balance += amount
		}

		// Update saldo dompet
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		// Simpan riwayat transaksi mutasi
		transaction := &entity.Transaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        txType,
			Description: desc,
		}
		return tx.Create(transaction).Error
	})
}
