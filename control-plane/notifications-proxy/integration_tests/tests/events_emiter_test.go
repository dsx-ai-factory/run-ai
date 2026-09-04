//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Event struct {
	EventType  string                 `json:"event_type"`
	TenantId   string                 `json:"tenant_id"`
	Properties map[string]interface{} `json:"properties"`
}

func RunEventsEmitter(port int, tenant string, times int, data map[string]interface{}) {
	host := GetEnvOrDefault("NOTIFY_HOST", "0.0.0.0")
	url := fmt.Sprintf("http://%s:%d/notify", host, port)

	data["tenant_id"] = tenant
	e := Event{
		EventType:  "email",
		TenantId:   tenant,
		Properties: data,
	}
	jsonData, err := json.Marshal(e)
	if err != nil {
		return
	}

	for i := 0; i < times; i += 1 {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		client := &http.Client{}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("failed to send event %+v\n", err)
			return
		}
		resp.Body.Close()
		fmt.Println("event sent")
	}
}
