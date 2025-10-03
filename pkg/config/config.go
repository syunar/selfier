// Package config
package config

import (
	"fmt"
	"strings"
	"time" // ADDED: For duration parsing

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload" // Ensures .env is loaded on import
	"github.com/spf13/viper"
)

// Config holds all the configuration for the application.
// Values are loaded by Viper from a config file and/or environment variables.
type Config struct {
	Primary  Primary        `mapstructure:"primary" validate:"required"`
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Auth     AuthConfig     `mapstructure:"auth" validate:"required"`
	AWS      AWSConfig      `mapstructure:"aws" validate:"required"`
	Logger   LoggerConfig   `mapstructure:"logger" validate:"required"`
}

type Primary struct {
	Env         string `mapstructure:"env" validate:"required,oneof=development production"`
	ServiceName string `mapstructure:"service_name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
}

func (p Primary) IsProd() bool {
	return p.Env == "production"
}

type ServerConfig struct {
	Port               string   `mapstructure:"port" validate:"required"`
	ReadTimeout        int      `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout       int      `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout        int      `mapstructure:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string           `mapstructure:"host" validate:"required"`
	Port            int              `mapstructure:"port" validate:"required"`
	User            string           `mapstructure:"user" validate:"required"`
	Password        string           `mapstructure:"password" validate:"required"`
	Name            string           `mapstructure:"name" validate:"required"`
	SSLMode         string           `mapstructure:"ssl_mode" validate:"required"`
	MaxOpenConns    int              `mapstructure:"max_open_conns" validate:"required"`
	MaxIdleConns    int              `mapstructure:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int              `mapstructure:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int              `mapstructure:"conn_max_idle_time" validate:"required"`
	GormLogger      GormLoggerConfig `mapstructure:"gorm_logger" validate:"required"`
}

// DSN returns the Data Source Name for connecting to the database.
func (db DatabaseConfig) DSN() string {
	// Example for PostgreSQL. Adjust for your specific database.
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		db.Host, db.User, db.Password, db.Name, db.Port, db.SSLMode)
}

// GormLoggerConfig for GORM-specific logging settings.
type GormLoggerConfig struct {
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" validate:"required"`
	IgnoreRecordNotFound bool          `mapstructure:"ignore_record_not_found"`
}

// LoggerConfig for application-wide logging settings.
type LoggerConfig struct {
	Level string `mapstructure:"level" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `mapstructure:"secret_key" validate:"required"`
}

type AWSConfig struct {
	Region           string `mapstructure:"region" validate:"required"`
	AccessKeyID      string `mapstructure:"access_key_id" validate:"required"`
	SecretAccessKey  string `mapstructure:"secret_access_key" validate:"required"`
	UploadBucket     string `mapstructure:"upload_bucket" validate:"required"`
	EndpointURL      string `mapstructure:"endpoint_url"`
	S3ForcePathStyle bool   `mapstructure:"s3_force_path_style"`
}

// LoadConfig reads configuration from file and/or environment variables.
func LoadConfig() (*Config, error) {
	var cfg Config
	viper.SetDefault("primary.env", "development")
	viper.SetDefault("primary.service_name", "selfier")
	viper.SetDefault("primary.version", "1.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("server.cors_allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "selfier")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", 5)
	viper.SetDefault("database.conn_max_idle_time", 5)
	viper.SetDefault("database.gorm_logger.slow_query_threshold", "200ms")
	viper.SetDefault("database.gorm_logger.ignore_record_not_found", false)
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("auth.secret_key", "your-secret-key")
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.access_key_id", "your-access-key-id")
	viper.SetDefault("aws.secret_access_key", "your-secret-access-key")
	viper.SetDefault("aws.upload_bucket", "your-upload-bucket")
	viper.SetDefault("aws.endpoint_url", "https://s3.amazonaws.com")
	viper.SetDefault("aws.s3_force_path_style", false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}
