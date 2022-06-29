package logging

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

//Creates a logger
func InitLogger() {
	zaplogger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize Zap logger")
	}
	defer zaplogger.Sync()
	sugar := zaplogger.Sugar()
	logger = sugar
}

//Returns our logger
func GetLogger() *zap.SugaredLogger {
	return logger
}
