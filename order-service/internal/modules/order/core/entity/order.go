package entity

import "time"

type Order struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             uint64    `gorm:"not null" json:"user_id"`
	DriverID           *uint64   `json:"driver_id"` // Nullable jika belum dicocokkan
	OriginLat          float64   `gorm:"type:decimal(9,6);not null" json:"origin_lat"`
	OriginLng          float64   `gorm:"type:decimal(9,6);not null" json:"origin_lng"`
	DestinationLat     float64   `gorm:"type:decimal(9,6);not null" json:"destination_lat"`
	DestinationLng     float64   `gorm:"type:decimal(9,6);not null" json:"destination_lng"`
	Fare               float64   `gorm:"type:decimal(12,2);not null" json:"fare"`
	Status             string    `gorm:"type:varchar(20);default:'created';not null" json:"status"` // created, matched, accepted, completed, cancelled
	UserNameSnapshot   string    `gorm:"type:varchar(100)" json:"user_name_snapshot"`               // Denormalisasi data untuk optimasi query
	UserPhoneSnapshot  string    `gorm:"type:varchar(20)" json:"user_phone_snapshot"`
	DriverNameSnapshot string    `gorm:"type:varchar(100)" json:"driver_name_snapshot"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type OrderRoute struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   uint64 `gorm:"uniqueIndex;not null" json:"order_id"`
	RoutePoly string `gorm:"type:text;not null" json:"route_poly"` // Menyimpan rute jemput-antar terenkripsi
}
