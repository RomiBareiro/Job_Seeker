package setup

import (
	"context"
	"fmt"
	"os"
	"strconv"

	d "jobs/db"
	"jobs/types"

	"github.com/go-playground/validator"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type flags struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	DBName   string `validate:"required"`
	SSLMode  string
}

// Setup
func Setup(ctx context.Context) (*d.DBConnector, error) {
	dbConfig, err := setupFlags()
	if err != nil {
		return nil, fmt.Errorf("could not get DB params: %w", err)
	}
	db, err := SetupDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("could not configure DB: %w", err)
	}

	logger, err := SetupLogger()
	if err != nil {
		return nil, fmt.Errorf("could not configure logger: %w", err)
	}
	return &d.DBConnector{DB: db, Logger: logger}, err
}

func setupFlags() (types.DatabaseConfig, error) {
	args := flags{
		Host: os.Getenv("POSTGRES_HOST"),
		Port: func() int {
			port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
			if err != nil {
				return 0
			}
			return port
		}(),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}

	v := validator.New()
	if err := v.Struct(args); err != nil {
		return types.DatabaseConfig{}, err
	}

	return types.DatabaseConfig{
		Host:     args.Host,
		Port:     args.Port,
		User:     args.User,
		Password: args.Password,
		DBName:   args.DBName,
		SSLMode:  args.SSLMode,
	}, nil
}

func SetupDB(config types.DatabaseConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetupLogger all necessary stuff to configure logger
func SetupLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
