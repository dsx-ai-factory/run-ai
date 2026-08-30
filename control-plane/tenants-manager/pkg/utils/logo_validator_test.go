package utils_test

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLogoIfValid_HappyFlow(t *testing.T) {
	logoValidatorMock := mocks.NewLogoValidator(t)
	logo := []byte("test")
	logoValidatorMock.On("DecodeLogo", mock.Anything).Return(&logo, nil)
	logoValidatorMock.On("ValidateLogoSize", mock.Anything).Return(nil)
	logoValidatorMock.On("ValidateMimeType", mock.Anything).Return(nil)

	result, err := utils.GetLogoIfValid(logoValidatorMock, "stam")
	assert.Nil(t, err)
	assert.Equal(t, &logo, result)
}

func TestGetLogoIfValid_FailBigSize(t *testing.T) {
	logoValidatorMock := mocks.NewLogoValidator(t)
	logo := []byte("test")
	expApiError := utils.Error{Code: 123, Message: "stam", Err: errors.New("stam")}
	logoValidatorMock.On("DecodeLogo", mock.Anything).Return(&logo, nil)
	logoValidatorMock.On("ValidateLogoSize", mock.Anything).Return(&expApiError)

	result, err := utils.GetLogoIfValid(logoValidatorMock, "stam")
	assert.Equal(t, &expApiError, err)
	assert.Nil(t, result)
}

func TestGetLogoIfValid_FailMimeType(t *testing.T) {
	logoValidatorMock := mocks.NewLogoValidator(t)
	logo := []byte("test")
	expApiError := utils.Error{Code: 123, Message: "stam", Err: errors.New("stam")}
	logoValidatorMock.On("DecodeLogo", mock.Anything).Return(&logo, nil)
	logoValidatorMock.On("ValidateLogoSize", mock.Anything).Return(nil)
	logoValidatorMock.On("ValidateMimeType", mock.Anything).Return(&expApiError)

	result, err := utils.GetLogoIfValid(logoValidatorMock, "stam")
	assert.Equal(t, &expApiError, err)
	assert.Nil(t, result)
}
