package util

import (
	"errors"
	"fmt"

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

func LogInfoln(log ...interface{}) {
	if instance == nil {
		fmt.Println(errors.New("instance of logger is nil (may be it was not initialized?)").Error())
		return
	}

	instance.Infoln(log)
}
