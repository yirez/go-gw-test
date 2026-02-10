package types

// StandardConfig captures shared settings used to start each service.
type StandardConfig struct {
	Env  string    `mapstructure:"env"`
	Port int       `mapstructure:"port"`
	DB   *DBConfig `mapstructure:"db"`
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
