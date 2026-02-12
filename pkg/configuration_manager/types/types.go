package types

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StandardConfig captures shared settings used to start each service.
type StandardConfig struct {
	Env         string          `mapstructure:"env"`
	Port        int             `mapstructure:"port"`
	DBConfig    *DBConfig       `mapstructure:"db"`
	AuthConfig  *AuthConfig     `mapstructure:"auth"`
	Clients     StandardClients `mapstructure:"-"`
	RedisConfig *RedisConfig    `mapstructure:"redis"`
}

// StandardClients provides shared service clients from configuration init.
type StandardClients struct {
	Logger *zap.Logger
	DB     *gorm.DB
	Redis  *redis.Client
}

// DBConfig captures database configuration for GORM.
type DBConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"sslmode"`
	TimeZone        string `mapstructure:"timezone"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime_sec"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time_sec"`
}

// RedisConfig captures redis connection details.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig captures auth service connection details.
type AuthConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	ServiceID string `mapstructure:"service_id"`
	Secret    string `mapstructure:"secret"`
}

// InitChecklist controls which standard clients should be initialized.
type InitChecklist struct {
	DB              bool
	Auth            bool
	Redis           bool
	AutoMigrateList []any
}
