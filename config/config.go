package config

import (
	"context"
	"fmt"
	"os"
	//"time"


	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver for PostgreSQL
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"database/sql"

	"github.com/pressly/goose/v3"
)

var (
	DB  *pgxpool.Pool
	Log *zap.Logger
)

func Init() {
	initLogger()
	initDB()
	runMigrations() // Run database migrations using Goose
}

func initDB() {
	var err error

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		Log.Fatal("Unable to parse database configuration", zap.Error(err))
	}

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		Log.Fatal("Unable to connect to database", zap.Error(err))
	}
}

func initLogger() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

func runMigrations() {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	//db, err := goose.OpenDBWithDriver("postgres", dsn)
	db, err := sql.Open("pgx", dsn) // Use the "pgx" driver here
	if err != nil {
		Log.Fatal("Failed to connect to database for migrations", zap.Error(err))
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		Log.Fatal("Failed to run migrations", zap.Error(err))
	}

	Log.Info("Database migrations applied successfully")
}