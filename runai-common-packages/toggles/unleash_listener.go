/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import unleash "github.com/Unleash/unleash-client-go/v3"

type unleashListener struct{}

func (l *unleashListener) OnError(err error) {
	log.Errorf("unleash error: %v", err)
}

func (l *unleashListener) OnWarning(warning error) {
	log.Warnf("unleash warning: %v", warning)
}

func (l *unleashListener) OnReady() {
	log.Infof("unleash ready")
}

func (l *unleashListener) OnCount(name string, enabled bool) {
	log.Debugf("unleash counted '%s' as enabled: %v", name, enabled)
}

func (l *unleashListener) OnSent(payload unleash.MetricsData) {
	log.Debugf("unleash sent metrics: %+v", payload)
}

func (l *unleashListener) OnRegistered(payload unleash.ClientData) {
	log.Debugf("unleash client registered: %+v", payload)
}
