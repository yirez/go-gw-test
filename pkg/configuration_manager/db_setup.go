package configuration_manager

import (
	"fmt"
	"go-gw-test/pkg/configuration_manager/types"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// initDB initializes a GORM Postgres connection if dbConfig is provided.
func initDB(dbConfig *types.DBConfig) (*gorm.DB, error) {
	if dbConfig == nil {
		err := fmt.Errorf("db config missing")
		log.Printf("init db: %v", err)
		return nil, err
	}

	dsn, err := buildPostgresDSN(dbConfig)
	if err != nil {
		log.Printf("init db dsn: %v", err)
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		err = fmt.Errorf("open gorm db: %w", err)
		log.Printf("init db open: %v", err)
		return nil, err
	}

	err = configureDBPool(db, dbConfig)
	if err != nil {
		log.Printf("init db pool: %v", err)
		return nil, err
	}

	return db, nil
}

func autoMigrateDB(db *gorm.DB, entities []any) error {
	if db == nil {
		err := fmt.Errorf("db not initialized")
		log.Printf("auto migrate: %v", err)
		return err
	}

	if len(entities) == 0 {
		return nil
	}

	err := db.AutoMigrate(entities...)
	if err != nil {
		err = fmt.Errorf("auto migrate: %w", err)
		log.Printf("auto migrate: %v", err)
		return err
	}

	return nil
}

// buildPostgresDSN builds a Postgres DSN for GORM.
func buildPostgresDSN(dbConfig *types.DBConfig) (string, error) {
	if strings.TrimSpace(dbConfig.Host) == "" {
		err := fmt.Errorf("db.host is required")
		log.Printf("build dsn: %v", err)
		return "", err
	}
	if dbConfig.Port == 0 {
		err := fmt.Errorf("db.port is required")
		log.Printf("build dsn: %v", err)
		return "", err
	}
	if strings.TrimSpace(dbConfig.Name) == "" {
		err := fmt.Errorf("db.name is required")
		log.Printf("build dsn: %v", err)
		return "", err
	}
	if strings.TrimSpace(dbConfig.User) == "" {
		err := fmt.Errorf("db.user is required")
		log.Printf("build dsn: %v", err)
		return "", err
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
func configureDBPool(db *gorm.DB, dbConfig *types.DBConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		err = fmt.Errorf("get sql db: %w", err)
		log.Printf("db pool: %v", err)
		return err
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
