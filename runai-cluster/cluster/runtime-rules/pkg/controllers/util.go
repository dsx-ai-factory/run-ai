package controllers

import (
	"go.uber.org/zap"
)

func genericControllerErrorHandler(err error, operation, resourceName string) {
	if err != nil {
		zap.S().Errorf("Had an error while trying to %s %s: %v ", operation, resourceName, err)
	}
}
