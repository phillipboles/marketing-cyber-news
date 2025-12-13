package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgxpool for database operations
type DB struct {
	Pool *pgxpool.Pool
}

// Config for database connection
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	MaxConns    int32
	MinConns    int32
	MaxConnLife time.Duration
	MaxConnIdle time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Host:        "localhost",
		Port:        5432,
		User:        "postgres",
		Password:    "",
		Database:    "aci_backend",
		SSLMode:     "disable",
		MaxConns:    25,
		MinConns:    5,
		MaxConnLife: time.Hour,
		MaxConnIdle: 30 * time.Minute,
	}
}

// NewDB creates a new database connection pool
func NewDB(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("database host is required")
	}

	if cfg.Port <= 0 {
		return nil, fmt.Errorf("database port must be positive")
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("database name is required")
	}

	if cfg.MaxConns <= 0 {
		return nil, fmt.Errorf("max connections must be positive")
	}

	if cfg.MinConns < 0 {
		return nil, fmt.Errorf("min connections cannot be negative")
	}

	if cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf("min connections cannot exceed max connections")
	}

	connString := buildConnectionString(cfg)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLife
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdle

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// buildConnectionString constructs a PostgreSQL connection string from Config
func buildConnectionString(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)
}

// Close closes the database connection pool gracefully
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping checks database connectivity
func (db *DB) Ping(ctx context.Context) error {
	if db.Pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// BeginTx starts a database transaction
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// Stats returns connection pool statistics
func (db *DB) Stats() *pgxpool.Stat {
	if db.Pool == nil {
		return nil
	}
	return db.Pool.Stat()
}
