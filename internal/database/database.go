package database

import (
	"fmt"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true, // 禁用复数表名
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&model.ProjectModel{},
		&model.ContributeRecordModel{},
		&model.EventModel{},
		&model.RefundRecordModel{},
		&model.SettlementRecordModel{},
		&model.ProjectTeamModel{},
		&model.ProjectMilestoneModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
