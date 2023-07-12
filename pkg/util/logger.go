package util

import (
	"errors"

	"go.uber.org/zap"
)

var instance *zap.SugaredLogger

func InitLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	instance = logger.Sugar()
	return nil
}

func GetLogger() *zap.SugaredLogger {
	return instance
}

func LogInfoln(log ...interface{}) error {
	if instance == nil {
		return errors.New("instance of logger is nil (may be it was not initialized?)")
	}

	instance.Infoln(log)
	return nil
}
