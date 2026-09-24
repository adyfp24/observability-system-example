package entity

import "time"

type User struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	PhoneNumber string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone_number"`
	Password    string    `gorm:"type:varchar(255);not null" json:"-"`
	Role        string    `gorm:"type:varchar(20);default:'customer'" json:"role"` // customer, driver, admin
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Driver struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"profile"`
	VehicleType string    `gorm:"type:varchar(20);not null" json:"vehicle_type"` // car_premium, car_regular, bike
	IsOnline    bool      `gorm:"default:false" json:"is_online"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
