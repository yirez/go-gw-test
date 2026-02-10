package configuration_manager

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StandardConfig captures shared settings used to start each service.
type StandardConfig struct {
	Env  string    `hcl:"env"`
	Port int       `hcl:"port"`
	DB   *DBConfig `hcl:"db,block"`
}

// DBConfig captures database configuration for GORM.
type DBConfig struct {
	Host            string `hcl:"host"`
	Port            int    `hcl:"port"`
	Name            string `hcl:"name"`
	User            string `hcl:"user"`
	Password        string `hcl:"password"`
	SSLMode         string `hcl:"sslmode,optional"`
	TimeZone        string `hcl:"timezone,optional"`
	MaxOpenConns    int    `hcl:"max_open_conns,optional"`
	MaxIdleConns    int    `hcl:"max_idle_conns,optional"`
	ConnMaxLifetime int    `hcl:"conn_max_lifetime_sec,optional"`
	ConnMaxIdleTime int    `hcl:"conn_max_idle_time_sec,optional"`
}

// InitStandardConfigs loads env/port from configPath and initializes a zap logger and GORM connection.
func InitStandardConfigs(configPath string) (StandardConfig, *zap.Logger, *gorm.DB, error) {
	var cfg StandardConfig
	if configPath == "" {
		return cfg, nil, nil, fmt.Errorf("configPath is required")
	}

	if err := hclsimple.DecodeFile(configPath, nil, &cfg); err != nil {
		return StandardConfig{}, nil, nil, fmt.Errorf("decode config: %w", err)
	}

	logger, err := buildLogger(cfg.Env)
	if err != nil {
		return StandardConfig{}, nil, nil, err
	}

	db, err := initDB(cfg.DB)
	if err != nil {
		return StandardConfig{}, nil, nil, err
	}

	return cfg, logger, db, nil
}

// buildLogger creates a zap logger based on the environment.
func buildLogger(env string) (*zap.Logger, error) {
	if env == "prod" {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}

// initDB initializes a GORM Postgres connection if dbConfig is provided.
func initDB(dbConfig *DBConfig) (*gorm.DB, error) {
	if dbConfig == nil {
		return nil, nil
	}

	dsn, err := buildPostgresDSN(dbConfig)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open gorm db: %w", err)
	}

	if err := configureDBPool(db, dbConfig); err != nil {
		return nil, err
	}

	return db, nil
}

// buildPostgresDSN builds a Postgres DSN for GORM.
func buildPostgresDSN(dbConfig *DBConfig) (string, error) {
	if strings.TrimSpace(dbConfig.Host) == "" {
		return "", fmt.Errorf("db.host is required")
	}
	if dbConfig.Port == 0 {
		return "", fmt.Errorf("db.port is required")
	}
	if strings.TrimSpace(dbConfig.Name) == "" {
		return "", fmt.Errorf("db.name is required")
	}
	if strings.TrimSpace(dbConfig.User) == "" {
		return "", fmt.Errorf("db.user is required")
	}

	sslMode := dbConfig.SSLMode
	if strings.TrimSpace(sslMode) == "" {
		sslMode = "disable"
	}

	timeZone := dbConfig.TimeZone
	if strings.TrimSpace(timeZone) == "" {
		timeZone = "UTC"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		sslMode,
		timeZone,
	), nil
}

// configureDBPool applies connection pool settings to the underlying sql.DB.
func configureDBPool(db *gorm.DB, dbConfig *DBConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	if dbConfig.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	}

	if dbConfig.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	}

	if dbConfig.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)
	}

	if dbConfig.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.ConnMaxIdleTime) * time.Second)
	}

	return nil
}
