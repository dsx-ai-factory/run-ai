# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package kubernetes

is_service {
	input.kind = "Service"
}

is_deployment {
	input.kind = "Deployment"
}

is_daemonset {
	input.kind = "DaemonSet"
}

is_statefulset {
	input.kind = "StatefulSet"
}