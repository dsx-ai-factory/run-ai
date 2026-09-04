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