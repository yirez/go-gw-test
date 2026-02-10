package configuration_manager

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"go-gw-test/pkg/configuration_manager/types"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StandardConfig captures shared settings used to start each service.
type StandardConfig = types.StandardConfig

// DBConfig captures database configuration for GORM.
type DBConfig = types.DBConfig

// InitStandardConfigs loads env/port from configPath and initializes a zap logger and GORM connection.
func InitStandardConfigs(configPath string) (StandardConfig, *zap.Logger, *gorm.DB, error) {
	var cfg StandardConfig
	if configPath == "" {
		err := fmt.Errorf("configPath is required")
		log.Printf("init config: %v", err)
		return cfg, nil, nil, err
	}

	v, err := loadConfig(configPath)
	if err != nil {
		log.Printf("load config: %v", err)
		return StandardConfig{}, nil, nil, err
	}

	err = v.Unmarshal(&cfg)
	if err != nil {
		err = fmt.Errorf("decode config: %w", err)
		log.Printf("decode config: %v", err)
		return StandardConfig{}, nil, nil, err
	}

	logger, err := buildLogger(cfg.Env)
	if err != nil {
		log.Printf("build logger: %v", err)
		return StandardConfig{}, nil, nil, err
	}

	db, err := initDB(cfg.DB)
	if err != nil {
		log.Printf("init db: %v", err)
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

	err = configureDBPool(db, dbConfig)
	if err != nil {
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

// ReadCustomConfig decodes a config attribute or block into target using a dotted key path.
func ReadCustomConfig(keyPath string, target any) error {
	if keyPath == "" {
		err := fmt.Errorf("keyPath is required")
		log.Printf("read custom config: %v", err)
		return err
	}
	if target == nil {
		err := fmt.Errorf("target is required")
		log.Printf("read custom config: %v", err)
		return err
	}

	configPath := "config.hcl"
	v, err := loadConfig(configPath)
	if err != nil {
		log.Printf("read custom config load: %v", err)
		return err
	}

	if v.IsSet(keyPath) {
		value := v.Get(keyPath)
		return decodeValueIntoTarget(value, target)
	}

	sub := v.Sub(keyPath)
	if sub == nil {
		err = fmt.Errorf("config key not found: %s", keyPath)
		log.Printf("read custom config: %v", err)
		return err
	}

	err = sub.Unmarshal(target)
	if err != nil {
		err = fmt.Errorf("decode config %s: %w", keyPath, err)
		log.Printf("read custom config decode: %v", err)
		return err
	}

	return nil
}

// loadConfig reads config.hcl into a viper instance.
func loadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("hcl")

	err := v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("read config: %w", err)
		log.Printf("load config: %v", err)
		return nil, err
	}

	return v, nil
}

// decodeValueIntoTarget maps a config value into the target reference.
func decodeValueIntoTarget(value any, target any) error {
	if value == nil {
		return fmt.Errorf("config value is nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	kind := targetValue.Elem().Kind()
	if kind == reflect.String {
		strValue, ok := value.(string)
		if ok {
			targetValue.Elem().SetString(strValue)
			return nil
		}
	}

	if kind == reflect.Int || kind == reflect.Int64 {
		intValue, ok := value.(int)
		if ok {
			targetValue.Elem().SetInt(int64(intValue))
			return nil
		}
		floatValue, ok := value.(float64)
		if ok {
			targetValue.Elem().SetInt(int64(floatValue))
			return nil
		}
	}

	if kind == reflect.Bool {
		boolValue, ok := value.(bool)
		if ok {
			targetValue.Elem().SetBool(boolValue)
			return nil
		}
	}

	decoderConfig := &mapstructure.DecoderConfig{
		Result:  target,
		TagName: "mapstructure",
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("build decoder: %w", err)
	}

	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("decode value: %w", err)
	}

	return nil
}
