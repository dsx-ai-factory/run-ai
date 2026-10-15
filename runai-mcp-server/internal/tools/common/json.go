/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ReadJSONResponse decodes a generated API client's buffered response body.
func ReadJSONResponse(resp *http.Response, out any) error {
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("missing response body")
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}
