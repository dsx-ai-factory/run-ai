/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const slackAuthCode = "5439c659-19a5-5335-ca5b-bfbc83b1fd77"

func main() {
	fmt.Println("Slack mock starting.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, os.Interrupt)

	fmt.Println("Slack mock started.")
	server := NewSlackMockServer(slackAuthCode)
	go server.Start()

	fmt.Println("Slack mock running.")
	<-signalChan
	fmt.Println("Received termination, shutting down.")
	server.Close()
}
