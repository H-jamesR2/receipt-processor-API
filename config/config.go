// config/config.go

package config

import (
	"fmt"
	"os"


	"rcpt-proc-challenge-ans/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	Log *zap.Logger
)

func Init() {
	initLogger()
	initDB()
}

func initDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_PASSWORD"))

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	err = DB.AutoMigrate(&model.Receipt{}, &model.Item{})
	if err != nil {
		Log.Fatal("Failed to migrate database", zap.Error(err))
	}
}

func initLogger() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

