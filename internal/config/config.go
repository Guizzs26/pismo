package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	AppPort  string
	LogLevel string
	DB       DBConfig
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheck     time.Duration
	ConnString      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}

	db := DBConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5768"),
		User:            getEnv("DB_USER", "pismousr"),
		Password:        getEnv("DB_PASSWORD", "pismopass"),
		Name:            getEnv("DB_NAME", "pismodb"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxConns:        getEnv[int32]("DB_MAX_CONNS", 25),
		MinConns:        getEnv[int32]("DB_MIN_CONNS", 5),
		MaxConnLifetime: time.Duration(getEnv("DB_MAX_CONN_LIFETIME_MIN", 30)) * time.Minute,
		MaxConnIdleTime: time.Duration(getEnv("DB_MAX_CONN_IDLE_TIME_MIN", 10)) * time.Minute,
		HealthCheck:     time.Duration(getEnv("DB_HEALTH_CHECK_SEC", 30)) * time.Second,
	}

	db.ConnString = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)

	return &Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		AppPort:  getEnv("APP_PORT", "9090"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		DB:       db,
	}
}

type envParseable interface {
	~string | ~int | ~int32 | ~int64 | ~float64 | ~bool
}

func getEnv[T envParseable](key string, fallback T) T {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}

	var result any
	var err error

	switch any(fallback).(type) {
	case string:
		result = v
	case int:
		result, err = strconv.Atoi(v)
	case int32:

		var i int64
		i, err = strconv.ParseInt(v, 10, 32)
		result = int32(i)
	case int64:
		result, err = strconv.ParseInt(v, 10, 64)
	case float64:
		result, err = strconv.ParseFloat(v, 64)
	case bool:
		result, err = strconv.ParseBool(v)
	default:
		return fallback
	}
	if err != nil {
		log.Printf("Invalid value for environment variable %s=%q, "+
			"using value defined in fallback: %v\n",
			key, v, fallback)
		return fallback
	}

	return result.(T)
}
