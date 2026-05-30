package database

import (
	"fmt"
	"time"

	"github.com/radius/radius-backend/internal/shared/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Postgres struct {
	DB     *gorm.DB
	logger *zap.Logger
}

func NewPostgres(cfg config.DatabaseConfig, log *zap.Logger) (*Postgres, error) {
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.SSLMode == "disable" {
		gormLogger = logger.Default.LogMode(logger.Warn)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:      gormLogger,
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	log.Info("connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)

	return &Postgres{DB: db, logger: log}, nil
}

func (p *Postgres) Close() {
	sqlDB, err := p.DB.DB()
	if err != nil {
		p.logger.Error("failed to get sql.DB for close", zap.Error(err))
		return
	}
	p.logger.Info("closing PostgreSQL connection pool")
	sqlDB.Close()
}

func (p *Postgres) HealthCheck() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
