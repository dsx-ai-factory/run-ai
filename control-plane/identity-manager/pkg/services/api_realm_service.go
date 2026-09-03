/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type RealmApiService struct {
	iamServices      iam.IamServiceContainer
	dal              business.RealmDal
	config           *config.TenantRealm
	masterReconciler reconciler.MasterReconcilerI
}

func NewRealmApiService(iamServices iam.IamServiceContainer, config *config.TenantRealm, masterReconciler *reconciler.MasterReconcilerI, dal business.RealmDal) api.RealmAPIServicer {
	return &RealmApiService{
		iamServices:      iamServices,
		config:           config,
		dal:              dal,
		masterReconciler: *masterReconciler,
	}
}

func (s *RealmApiService) CreateRealm(ctx context.Context, realmCreationRequest api.RealmCreationRequest) (api.ImplResponse, error) {
	_, err := s.dal.CreateRealm(ctx, &models.Realm{
		TenantId: realmCreationRequest.TenantId,
		Realm:    realmCreationRequest.RealmName,
	})
	if err != nil {
		if utils.IsConflict(err) {
			msg := fmt.Sprintf("Resource with the name '%s' already exists", realmCreationRequest.RealmName)
			return api.Response(http.StatusConflict, utils.Error{Code: http.StatusConflict, Message: msg}), utils.Error{Code: http.StatusConflict, Message: msg}
		}
		errMsg := fmt.Sprintf(`failed creating realm with name %s.`, realmCreationRequest.RealmName)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, err: %v", errMsg, err)
		return utils.HandleError(ctx, "realm", "realm name", realmCreationRequest.RealmName, err, errMsg)
	}
	defaults := &iamModels.RealmDefaults{
		Theme:         &s.config.LoginTheme,
		SmtpHost:      s.config.SmtpHost,
		SmtpPort:      s.config.SmtpPort,
		SmtpFromEmail: s.config.SmtpFromEmail,
	}
	realmModel := iamModels.NewRunaiRealm(realmCreationRequest.RealmName, defaults)
	_, err = s.iamServices.RealmService.CreateRealm(ctx, realmModel)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating realm with name %s.`, realmCreationRequest.RealmName)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, err: %v", errMsg, err)
		return utils.HandleError(ctx, "realm", "realm name", realmCreationRequest.RealmName, err, errMsg)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Realm %v created, reconciling", realmCreationRequest.RealmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, realmCreationRequest.TenantId)

	return s.reconcileRelmAndBuildResponse(ctx, realmCreationRequest.RealmName, http.StatusOK)
}

func (s *RealmApiService) GetOrCreateRealm(ctx context.Context, realmCreationRequest api.RealmCreationRequest) (api.ImplResponse, error) {
	implResp, err := s.CreateRealm(ctx, realmCreationRequest)
	if err != nil {
		if implResp.Code != http.StatusConflict {
			return implResp, err
		}
		existing, getErr := s.dal.GetRealmByName(ctx, realmCreationRequest.RealmName)
		if getErr != nil {
			return implResp, getErr
		}
		if existing.TenantId != realmCreationRequest.TenantId {
			msg := fmt.Sprintf("realm '%s' already exists for a different tenant", realmCreationRequest.RealmName)
			return api.Response(http.StatusConflict, utils.Error{Code: http.StatusConflict, Message: msg}), utils.Error{Code: http.StatusConflict, Message: msg}
		}
		ctx = context.WithValue(ctx, runaictx.CfTenantId, realmCreationRequest.TenantId)
		return s.reconcileRelmAndBuildResponse(ctx, realmCreationRequest.RealmName, http.StatusOK)
	}
	implResp.Code = http.StatusCreated
	return implResp, nil
}

func (s *RealmApiService) reconcileRelmAndBuildResponse(ctx context.Context, realmName string, statusCode int) (api.ImplResponse, error) {
	err := s.masterReconciler.Reconcile(ctx, realmName, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile realm %v, err: %v", realmName, err)
		return api.Response(http.StatusServiceUnavailable, err.Error()), utils.Error{Code: http.StatusServiceUnavailable, Message: err.Error()}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Realm %v reconciled", realmName)

	respClients, err := s.getClientsResponse(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get clients for realm %v, err: %v", realmName, err)
		return api.Response(http.StatusInternalServerError, err.Error()), utils.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	domain, err := utils.TemplateBaseUrl(s.config.TenantDomainTemplate, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to template base url with host template %s and realm %s: %v", s.config.TenantDomainTemplate, realmName, err)
		return api.Response(http.StatusInternalServerError, err.Error()), utils.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	resp := api.RealmCreationResponse{
		RealmName: realmName,
		Domain:    domain,
		Clients:   respClients,
	}
	return api.Response(statusCode, resp), nil
}

func (s *RealmApiService) DeleteRealm(ctx context.Context, realmName string, isHardDelete bool) (api.ImplResponse, error) {
	var err error
	if isHardDelete {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("hard deleting realm: %s", realmName)
		err = s.deleteRealm(ctx, realmName)
		if err != nil {
			errMsg := fmt.Sprintf(`failed delete realm with name %s.`, realmName)
			return utils.HandleError(ctx, "realm", "realm name", realmName, err, errMsg)
		}
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("soft deleting realm: %s", realmName)
		err = s.disableRealm(ctx, realmName)
		if err != nil {
			errMsg := fmt.Sprintf(`failed disable realm with name %s.`, realmName)
			return utils.HandleError(ctx, "realm", "realm name", realmName, err, errMsg)
		}
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *RealmApiService) ReconcileRealm(ctx context.Context, realmName string, request api.RealmReconcileRequest) (api.ImplResponse, error) {
	reconcilerContext := context.Background()
	reconcilerContext = context.WithValue(reconcilerContext, utils.AuthRealmKey, realmName)
	reconcilerContext = context.WithValue(reconcilerContext, runaictx.CfTenantId, runaictx.GetString(ctx, runaictx.CfTenantId))
	if request.ReconcilerType == nil {
		go s.launchAllReconcilers(reconcilerContext, realmName, request.Force)
		return api.Response(http.StatusAccepted, nil), nil
	}
	err := s.launchReconcileByType(ctx, realmName, request.Force, *request.ReconcilerType)
	if err != nil {
		errMsg := fmt.Sprintf(`failed running reconciler by type: %s for realm: %s.`, *request.ReconcilerType, realmName)
		return utils.HandleError(ctx, "realm", "realm name", realmName, err, errMsg)
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *RealmApiService) getClientsResponse(ctx context.Context, realm string) ([]api.Client, error) {
	clientsIds := []string{s.config.DefaultClientId, utils.RunaiInstallerClientId}
	respClients := []api.Client{}
	for _, clientId := range clientsIds {
		client, err := s.iamServices.ClientService.GetClientByClientId(ctx, realm, clientId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get client %v, err: %v", realm, err)
			return nil, err
		}
		if client == nil || client.ID == nil || client.ClientID == nil {
			err := fmt.Errorf("client %s in realm %s is missing required fields", clientId, realm)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%v", err)
			return nil, err
		}
		secret := ""
		if client.Secret != nil {
			secret = *client.Secret
		} else {
			secret, err = s.iamServices.ClientService.GetClientSecret(ctx, realm, *client.ID)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get client secret %v, err: %v", clientId, err)
				return nil, err
			}
		}
		respClients = append(respClients, api.Client{
			Id:       *client.ID,
			ClientId: *client.ClientID,
			Secret:   secret,
		})
	}
	return respClients, nil
}

func (s *RealmApiService) deleteRealm(ctx context.Context, realmName string) error {
	_, err := s.dal.DeleteRealm(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete realm %v from db, err: %v", realmName, err)
	}
	err = s.iamServices.RealmService.DeleteRealm(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete realm %v, err: %v", realmName, err)
		return err
	}
	return nil
}

func (s *RealmApiService) disableRealm(ctx context.Context, realmName string) error {
	_, err := s.dal.DisableRealm(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete realm %v from db, err: %v", realmName, err)
	}
	err = s.iamServices.RealmService.DisableRealm(ctx, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to disable realm %v, err: %v", realmName, err)
		return err
	}
	return nil
}

func (s *RealmApiService) launchAllReconcilers(ctx context.Context, realmName string, force bool) error {
	return s.masterReconciler.Reconcile(ctx, realmName, force)
}

func (s *RealmApiService) launchReconcileByType(ctx context.Context, realmName string, force bool, reconcilerType api.ReconcilerType) error {
	reconciler := s.masterReconciler.GetReconcilerByType(reconcilerType)
	if reconciler == nil {
		return fmt.Errorf("reconciler type: %s is not supported", reconcilerType)
	}
	return reconciler.Reconcile(ctx, realmName, force)
}
