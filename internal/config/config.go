package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

func NewConfig() (*Config, error) {

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env: %w", err)
	}

	return &cfg, nil
}

type Config struct {
	HTTP   HTTPConfig
	DB     DBConfig
	Logger LoggerConfig
}

type LoggerConfig struct {
	Level string `env:"LOG_LEVEL" env-default:"info"`
}

type HTTPConfig struct {
	Port string `env:"APP_PORT" env-default:"8080"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     string `env:"DB_PORT" env-required:"true"`
	User     string `env:"DB_USER" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME" env-required:"true"`
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", d.User, d.Password, d.Host, d.Port, d.Name)
}
