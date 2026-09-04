/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	daoMocks "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/notify/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	smocks "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestGetSink(t *testing.T) {
	requestedSinkId := "email1"

	dao := daoMocks.NewSinksDao(t)
	dao.On("FetchSink", requestedSinkId).Return(&models.SinkConfig{
		SinkId: "email1",
		Type:   "email",
		Config: datatypes.JSON(`{
				"type": "email",
				"password": "password must be removed"
			}`),
	}, nil)

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.GetSink(context.Background(), requestedSinkId)
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
	require.NotNil(t, res.Body)

	body := res.Body.(*api.SinkData)
	require.Equal(t, "email", body.Type)
	require.Equal(t, "email1", body.Id)
	require.Nil(t, body.Config["password"])
}

func TestGetSinks_All(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("FetchAllSinks").Return([]*models.SinkConfig{
		{
			SinkId: "email1",
			Type:   "email",
			Config: datatypes.JSON(`{
				"type": "email",
				"password": "password must be removed"
			}`),
		},
	}, nil)

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.GetSinks(context.Background())
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
	require.NotNil(t, res.Body)

	body := res.Body.([]api.SinkData)
	require.Len(t, body, 1)
	require.Equal(t, "email", body[0].Type)
	require.Equal(t, "email1", body[0].Id)
	require.Nil(t, body[0].Config["password"])
}

func TestUpdateSink_SucceedPasswordNotUpdated(t *testing.T) {
	ep, _ := encrypt.Encrypt("stam")
	sinkFromAPI := anyMap{
		"user": "updatedUserNoPass@mail.com",
	}
	sinkFromDB := datatypes.JSON(fmt.Sprintf(`{"password":"%s","user":"origUserWithPass@mail.com"}`, ep))
	mergedSink := datatypes.JSON(fmt.Sprintf(`{"password":"%s","user":"updatedUserNoPass@mail.com"}`, ep))

	res, err := updateSinkMocked(daoMocks.NewSinksDao(t), sinkFromAPI, sinkFromDB, mergedSink)

	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestUpdateSink_SucceedPasswordUpdated(t *testing.T) {
	oldPass, _ := encrypt.Encrypt("oldPassword")
	newPass, _ := encrypt.Encrypt("newPassword")
	sinkFromAPI := anyMap{
		"user":     "user@mail.com",
		"password": newPass,
	}
	sinkFromDB := datatypes.JSON(fmt.Sprintf(`{"password":"%s","user":"user@mail.com"}`, oldPass))
	mergedSink := datatypes.JSON(fmt.Sprintf(`{"password":"%s","user":"user@mail.com"}`, newPass))

	res, err := updateSinkMocked(daoMocks.NewSinksDao(t), sinkFromAPI, sinkFromDB, mergedSink)

	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestUpdateSink_NotFound(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("FetchSink", "missing-sink").Return((*models.SinkConfig)(nil), nil)

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}
	req := api.SinkDataForUpdate{
		Type:   "email",
		Config: anyMap{"user": "u@mail.com"},
	}

	res, err := sap.UpdateSink(context.Background(), "missing-sink", req)

	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "sink not found", apiErr.Message)
	dao.AssertExpectations(t)
}

func TestUpdateSink_FailedToFetch(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("FetchSink", "some-sink").Return((*models.SinkConfig)(nil), fmt.Errorf("failed to fetch sink"))

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}
	req := api.SinkDataForUpdate{
		Type:   "email",
		Config: anyMap{"user": "u@mail.com"},
	}

	res, err := sap.UpdateSink(context.Background(), "some-sink", req)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	dao.AssertExpectations(t)
}

func TestUpdateSink_FailedToUpdate(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("FetchSink", "some-sink").Return(
		&models.SinkConfig{
			SinkId: "some-sink",
			Config: datatypes.JSON(`{"user": "u@mail.com"}`),
		}, nil)
	dao.On("UpdateSink", "some-sink", "email", mock.Anything).Return(fmt.Errorf("failed to update sink"))

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}
	req := api.SinkDataForUpdate{
		Type:   "email",
		Config: anyMap{"user": "user@mail.com"},
	}

	res, err := sap.UpdateSink(context.Background(), "some-sink", req)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	dao.AssertExpectations(t)
}

func updateSinkMocked(dao *daoMocks.SinksDao, sinkFromAPI anyMap, sinkFromDB, mergedSink datatypes.JSON) (api.ImplResponse, error) {
	dao.On("FetchSink", "email1", mock.Anything).Return(
		&models.SinkConfig{
			SinkId: "email1",
			Config: sinkFromDB,
		}, nil)
	dao.On("UpdateSink", "email1", "email", mergedSink).Return(nil)

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.SinkDataForUpdate{
		Type:   "email",
		Config: sinkFromAPI,
	}

	return sap.UpdateSink(context.Background(), "email1", requestData)
}

func TestCreateSink_Success(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	req := api.SinkData{
		Id:     "sink1",
		Type:   "email",
		Config: anyMap{"user": "u@mail.com", "password": "pwd"},
	}
	expectedConfig, err := json.Marshal(req.Config)
	require.NoError(t, err)
	dao.On("CreateSink", "sink1", "email", datatypes.JSON(expectedConfig)).Return(nil)

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.CreateSink(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
	dao.AssertExpectations(t)
}

func TestCreateSink_DAOError(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("CreateSink", "sink1", "email", mock.Anything).Return(errors.New("db error"))

	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}
	req := api.SinkData{
		Id:     "sink1",
		Type:   "email",
		Config: anyMap{"user": "u@mail.com"},
	}

	res, err := sap.CreateSink(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	dao.AssertExpectations(t)
}

func TestDeleteSink(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("DeleteSink", "email1").Return(nil)
	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.DeleteSink(context.Background(), "email1")

	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestDeleteSink_Error_NotFound(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("DeleteSink", "email1").Return(configprovider.NewError("sink not found", configprovider.ErrCodeNotFound))
	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.DeleteSink(context.Background(), "email1")

	apiErr := err.(*ApiError)
	assert.Equal(t, "sink not found", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusNotFound)
}

func TestDeleteSink_Error_FailedToDelete(t *testing.T) {
	dao := daoMocks.NewSinksDao(t)
	dao.On("DeleteSink", "email1").Return(errors.New("failed to delete sink"))
	sap := SinksAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := sap.DeleteSink(context.Background(), "email1")

	apiErr := err.(*ApiError)
	assert.Equal(t, "failed to delete sink", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestSinkValidation_ValidCase(t *testing.T) {
	factoryMock := &mocks.Factory{}
	sinkMock := &smocks.Sinker{}
	factoryMock.On("CreateSinks", mock.Anything).Return([]sinks.Sinker{sinkMock}, nil)
	sinkMock.On("Verify").Return(nil)

	apiServices := SinksAPIService{
		sinkFactory: factoryMock,
	}
	response, err := apiServices.ValidateSink(context.Background(), createAValidValidationRequest())
	require.NoError(t, err)
	require.Equal(t, 200, response.Code)
	res := response.Body.(api.ValidateSink200Response)
	require.Equal(t, int32(ErrSuccess), res.ErrorCode)
}

func TestSinkValidation_InvalidCase(t *testing.T) {
	factoryMock := &mocks.Factory{}
	factoryMock.On("CreateSinks", mock.Anything).Return(nil, fmt.Errorf("failed creating a sink"))

	apiServices := SinksAPIService{
		sinkFactory: factoryMock,
	}
	req := createAValidValidationRequest()
	req.Id = "sinkId"
	response, err := apiServices.ValidateSink(context.Background(), req)
	require.Equal(t, 500, response.Code)
	apiErr := err.(*ApiError)
	require.Contains(t, apiErr.Message, "failed creating a sink")
}

func createAValidValidationRequest() api.SinkData {
	return api.SinkData{
		Id:   "email1",
		Type: "email",
		Config: anyMap{
			"type":     "email",
			"User":     "user",
			"Password": "password",
			"SmtpHost": "smtpHost",
			"SmtpPort": 111,
		},
	}
}
