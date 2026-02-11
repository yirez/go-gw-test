package configuration_manager

import (
	"fmt"
	"go-gw-test/pkg/configuration_manager/types"
	"log"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitStandardConfigs loads env/port from configPath, initializes standard clients, and optionally auto-migrates.
func InitStandardConfigs(initCheckList types.InitChecklist) (types.StandardConfig, error) {
	var cfg types.StandardConfig

	v, err := loadConfig("config.yml")
	if err != nil {
		log.Printf("load config: %v", err)
		return types.StandardConfig{}, err
	}

	err = v.Unmarshal(&cfg)
	if err != nil {
		err = fmt.Errorf("decode config: %w", err)
		log.Printf("decode config: %v", err)
		return types.StandardConfig{}, err
	}

	logger, err := buildLogger(cfg.Env)
	if err != nil {
		log.Printf("build logger: %v", err)
		return types.StandardConfig{}, err
	}
	cfg.Clients.Logger = logger

	var db *gorm.DB
	if initCheckList.DB {
		db, err = initDB(cfg.DBConfig)
		if err != nil {
			log.Printf("init db: %v", err)
			return types.StandardConfig{}, err
		}

		if cfg.Env != "prod" {
			err = autoMigrateDB(db, initCheckList.AutoMigrateList)
			if err != nil {
				log.Printf("auto migrate: %v", err)
				return types.StandardConfig{}, err
			}
		}
		cfg.Clients.DB = db
	}

	var redisClient *redis.Client
	if initCheckList.Redis {
		redisClient = newRedisClient(cfg.RedisConfig)
		if redisClient == nil {
			err = fmt.Errorf("redis not configured")
			log.Printf("init redis: %v", err)
			return types.StandardConfig{}, err
		}
		cfg.Clients.Redis = redisClient
	}

	return cfg, nil
}

// buildLogger creates a zap logger based on the environment.
func buildLogger(env string) (*zap.Logger, error) {
	if env == "prod" {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
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

	configPath := "config.yml"
	v, err := loadConfig(configPath)
	if err != nil {
		log.Printf("read custom config load: %v", err)
		return err
	}

	if v.IsSet(keyPath) {
		value := v.Get(keyPath)
		err = decodeValueIntoTarget(value, target)
		if err != nil {
			log.Printf("read custom config decode value: %v", err)
			return err
		}
		return nil
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

// loadConfig reads config.yml into a viper instance.
func loadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

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
		err := fmt.Errorf("config value is nil")
		log.Printf("decode value: %v", err)
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		err := fmt.Errorf("target must be a non-nil pointer")
		log.Printf("decode value: %v", err)
		return err
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
		err = fmt.Errorf("build decoder: %w", err)
		log.Printf("decode value: %v", err)
		return err
	}

	err = decoder.Decode(value)
	if err != nil {
		err = fmt.Errorf("decode value: %w", err)
		log.Printf("decode value: %v", err)
		return err
	}

	return nil
}
