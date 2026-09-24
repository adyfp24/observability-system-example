package config

import (
	"fmt"
	"log"
	"time"

	model "github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/entity"

	"github.com/adyfp24/okejek-go-service/order-service/internal/app/telemetry"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *Config) *gorm.DB {
	var (
		host     = cfg.Get("DB_HOST", "localhost")
		port     = cfg.Get("DB_PORT", "5432")
		user     = cfg.Get("DB_USER", "postgres")
		password = cfg.Get("DB_PASSWORD", "password")
		dbname   = cfg.Get("DB_NAME", "order")
		sslMode  = cfg.Get("DB_SSL_MODE", "disable")
		timeZone = cfg.Get("DB_TIMEZONE", "Asia/Jakarta")

		maxOpenConns = cfg.GetInt("DB_CONN_OPEN", 100)
		maxIdleConns = cfg.GetInt("DB_CONN_IDLE", 10)
		connLifeTime = cfg.GetInt("DB_CONN_LIFETIME", 15)
	)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, password, dbname, sslMode, timeZone)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	telemetry.InstrumentDB(db)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(connLifeTime))

	if err := db.AutoMigrate(&model.Order{}, &model.OrderRoute{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database connection established and migrations completed")

	return db
}
