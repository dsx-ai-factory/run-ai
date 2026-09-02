/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	errWrap "github.com/pkg/errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/retry"
)

type pubsubConnector struct {
	logger        logger.Loggerer
	pubSubFactory PubSubFactoryer
	errors        chan error
}

func newPubsubConnector(logger logger.Loggerer) *pubsubConnector {
	return &pubsubConnector{
		logger:        logger,
		pubSubFactory: newPubSubFactory(),
		errors:        make(chan error, 10),
	}
}

func (psc *pubsubConnector) Connect() error {
	conf := config.Get[*config_models.Pubsub]()
	wg := psc.connectPubSubers(context.Background(), conf.ConnectRetries+1)
	wg.Wait()

	if IsInitialized() {
		return nil
	}
	return getError(psc.errors)
}

func (psc *pubsubConnector) ConnectAsync(ctx context.Context) chan error {
	psc.connectPubSubers(ctx, math.MaxInt16)
	return psc.errors
}

func (psc *pubsubConnector) connectPubSubers(ctx context.Context, maxAttempts int) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go psc.connectPubSuber(ctx, false, &pubSuberInstance, maxAttempts, wg)

	conf := config.Get[*config_models.Pubsub]()
	initializeAdditional, err := initializeAdditionalPubsuber(conf)
	if err != nil {
		psc.logger.Warnf(err.Error())
	}
	if !initializeAdditional {
		return wg
	}

	psc.logger.Infof("connecting to additional pubsuber, type [%s] url [%s]", conf.AdditionalSubscriberType, conf.AdditionalSubscriberUrl)
	wg.Add(1)
	go psc.connectPubSuber(ctx, true, &additionalPubSuberInstance, maxAttempts, wg)

	return wg
}

func initializeAdditionalPubsuber(conf *config_models.Pubsub) (bool, error) {
	if !conf.EnableAdditionalSubscriber {
		return false, nil
	}

	if err := validateAdditionalSubscriberConfig(conf); err != nil {
		return false, fmt.Errorf("failed to validate additional subscriber config, err: %v", err)
	}
	return true, nil
}

func validateAdditionalSubscriberConfig(conf *config_models.Pubsub) error {
	if conf.AdditionalSubscriberType == "" {
		return fmt.Errorf("additional subscriber type is empty")
	}

	if conf.AdditionalSubscriberUrl == "" {
		return fmt.Errorf("additional subscriber url is empty")
	}

	if conf.AdditionalSubscriberUrl == conf.Url {
		return fmt.Errorf("additional subscriber url is the same as the main subscriber url")
	}

	return nil
}

func (psc *pubsubConnector) connectPubSuber(ctx context.Context, isAdditionalConnection bool, currentInstance *PubSuber, maxAttempts int, wg *sync.WaitGroup) {
	defer wg.Done()
	conf := config.Get[*config_models.Pubsub]()
	url, user, password, psType := conf.Url, conf.User, conf.Password, conf.Type
	if isAdditionalConnection {
		url, user, password, psType = conf.AdditionalSubscriberUrl, conf.AdditionalSubscriberUser, conf.AdditionalSubscriberPassword, conf.AdditionalSubscriberType
	}

	psi, err := psc.pubSubFactory.Create(psc.logger, psType)
	if err != nil {
		psc.logger.Errorf("failed to create pubsub instance, is_additional: %v err %+v", isAdditionalConnection, err)
		tryAppendToChan(psc.errors, errWrap.Wrapf(err, "failed to create pubsub instance, is_additional: %v", isAdditionalConnection))
		return
	}

	err = retry.Retry(ctx,
		psc.getRetryAttemptFunc(psi, isAdditionalConnection, url, user, password, maxAttempts),
		maxAttempts,
	)

	if err != nil {
		tryAppendToChan(psc.errors, errWrap.Wrapf(err, "failed to connect pubsub instance, is_additional: %v", isAdditionalConnection))
		return
	}

	select {
	case <-ctx.Done():
		psc.logger.Infof("context canceled, terminating")
		return
	default:
	}

	if *currentInstance != nil {
		_ = (*currentInstance).Close()
	}
	*currentInstance = psi
}

func (psc *pubsubConnector) getRetryAttemptFunc(psi PubSuber, isAdditionalConnection bool, url, user, password string, maxAttempts int) func(attempt uint) error {
	return func(attempt uint) error {
		if err := psi.Connect(url, user, password); err != nil {
			psc.logger.Warnf("Failed to connect to pubsub, additional: %v (attempt %d/%d): %v", isAdditionalConnection, attempt, maxAttempts, err)
			tryAppendToChan(psc.errors, errWrap.Wrapf(err, "failed to connect pubsub instance, is_additional: %v", isAdditionalConnection))
			return err
		}
		return nil
	}
}

func getError(errs chan error) error {
	var errMain, errAdditional error
	for {
		select {
		case err := <-errs:
			if strings.Contains(err.Error(), "is_additional: false") {
				errMain = err
			} else if strings.Contains(err.Error(), "is_additional: true") {
				errAdditional = err
			}
		default:
			return errors.Join(errMain, errAdditional)
		}
	}
}

func tryAppendToChan[T any](channel chan T, element T) {
	select {
	case channel <- element:
	default:
	}
}
