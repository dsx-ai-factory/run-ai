/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package network

import (
	"fmt"
	"net"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
)

const (
	lowestPort  = 1
	highestPort = 65535
)

func ValidAddress(address string) error {
	_, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("address %s is not in the format of host:port", address)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < lowestPort || port > highestPort {
		return fmt.Errorf("port %s is not a valid"+
			" port number should be between %d and %d", portStr, lowestPort, highestPort)
	}

	return nil
}

type addressCheck struct {
	address string
}

func (ac *addressCheck) Check() error {
	return ValidAddress(ac.address)
}

type AddressValidator struct {
	check addressCheck
}

func (av *AddressValidator) getFailMsg() string {
	return consts.InvalidAddress
}

func (av *AddressValidator) validate() error {
	return av.check.Check()
}

func (av *AddressValidator) Execute() (string, error) {
	if err := av.validate(); err != nil {
		return av.getFailMsg(), err
	}
	return "", nil
}

func NewAddressValidator(address string) *AddressValidator {
	return &AddressValidator{check: addressCheck{address: address}}
}
