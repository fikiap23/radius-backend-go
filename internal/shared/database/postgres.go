package database

import (
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/shared/config"
	"go.uber.org/zap"
)

type Postgres struct {
	Client *ent.Client
	logger *zap.Logger
	sqlDB  *sql.DB
}

func NewPostgres(cfg config.DatabaseConfig, log *zap.Logger) (*Postgres, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	log.Info("connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)

	return &Postgres{Client: client, logger: log, sqlDB: db}, nil
}

func (p *Postgres) Close() {
	p.logger.Info("closing PostgreSQL connection pool")
	if err := p.Client.Close(); err != nil {
		p.logger.Error("failed closing ent client", zap.Error(err))
	}
}

func (p *Postgres) HealthCheck() error {
	return p.sqlDB.Ping()
}
