package entity

import "time"

type Wallet struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex;not null" json:"user_id"` // Bisa ID customer atau driver
	Balance   float64   `gorm:"type:decimal(15,2);default:0.00;not null" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	WalletID    uint64    `gorm:"not null" json:"wallet_id"`
	Amount      float64   `gorm:"type:decimal(12,2);not null" json:"amount"` // Positif jika topup/pendapatan, negatif jika pembayaran/potongan
	Type        string    `gorm:"type:varchar(20);not null" json:"type"`     // credit (tambah), debit (potong)
	Description string    `gorm:"type:varchar(255)" json:"description"`      // e.g. "Absensi Harian Premium Car", "Pembayaran Order #88"
	CreatedAt   time.Time `json:"created_at"`
}
