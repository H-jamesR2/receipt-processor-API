package config

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func GenerateUUID() uuid.UUID {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		Log.Error("Failed to generate UUID", zap.Error(err))
		panic(err) // Handle according to your application's needs
	}
	Log.Info("Generated UUID", zap.String("uuid", newUUID.String()))
	return newUUID
}
