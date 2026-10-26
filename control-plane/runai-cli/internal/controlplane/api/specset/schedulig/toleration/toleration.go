/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toleration

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func NewList(list []string) ([]openapi.Toleration, error) {
	var objList = make([]openapi.Toleration, 0)
	for _, att := range list {
		obj, err := New(att)
		if err != nil {
			return nil, err
		}
		objList = append(objList, *obj)
	}
	return objList, nil
}

func New(att string) (*openapi.Toleration, error) {
	var (
		err error
	)

	p := NewParser(att)
	err = p.Parse()
	if err != nil {
		return nil, err
	}

	obj, err := create(p)
	if err != nil {
		return nil, err
	}

	err = validate(obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func create(am *parser.AttributeMapper) (*openapi.Toleration, error) {
	err := parser.CheckMandatoryFields(am, Operator, Key, Effect)
	if err != nil {
		return nil, err
	}

	op, err := openapi.NewTolerationOperatorFromValue(*am.Get(Operator))
	if err != nil {
		return nil, err
	}

	effect, err := openapi.NewTolerationEffectFromValue(*am.Get(Effect))
	if err != nil {
		return nil, err
	}

	toleration := &openapi.Toleration{
		Operator: *openapi.NewNullableTolerationOperator(op),
		Key:      *openapi.NewNullableString(am.Get(Key)),
		Effect:   *openapi.NewNullableTolerationEffect(effect),
	}

	if am.Get(Name) != nil {
		toleration.SetName(*am.Get(Name))
	}

	if am.Get(Value) != nil {
		toleration.SetValue(*am.Get(Value))
	}

	if am.Get(Seconds) != nil {
		sec, err := strconv.ParseInt(*am.Get(Seconds), 10, 32)
		if err != nil {
			return nil, err
		}
		toleration.SetSeconds(int32(sec))
	}

	return toleration, nil
}

func validate(toleration *openapi.Toleration) error {
	if toleration.Key.Get() == nil {
		return errors.New("tolerations: key must be provided")
	}

	if toleration.Operator.Get() == nil {
		return errors.New("tolerations: operator must be provided")
	}

	if toleration.Effect.Get() == nil {
		return errors.New("tolerations: effect must be provided")
	}

	if *toleration.Operator.Get() == openapi.TOLERATIONOPERATOR_EQUAL && !toleration.Value.IsSet() {
		return fmt.Errorf("tolerations: value must be provided when operator is set to %s", openapi.TOLERATIONOPERATOR_EQUAL)
	}

	if *toleration.Operator.Get() == openapi.TOLERATIONOPERATOR_EXISTS && toleration.Value.IsSet() {
		return fmt.Errorf("tolerations: value is not allowed when operator is set to %s", openapi.TOLERATIONOPERATOR_EXISTS)
	}

	if *toleration.Effect.Get() != openapi.TOLERATIONEFFECT_NO_EXECUTE && toleration.Seconds.IsSet() {
		return fmt.Errorf("tolerations: seconds is not allowed when effect is not set to %s", openapi.TOLERATIONEFFECT_NO_EXECUTE)
	}

	return nil
}
