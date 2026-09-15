{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Cluster Alert was triggered*"
    }
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "{{- range $key, $value := .EventProperties -}} *{{ $key }}:* {{ $value }}\n {{- end -}}"
    }
  }
]