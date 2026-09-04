[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*{{ .EventProperties.title }}*"
    }
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "{{- if .EventProperties.user_id -}} Hi {{.EventProperties.user_id }}, {{else}} Dear recipient, {{- end -}}"
    }
  },
  {
    "type": "section",
    "text": {
      "type": "plain_text",
      "text": {{ printf "%q" (or .EventProperties.description "") }}
    }
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "{{- range $key, $value := .EventProperties -}} {{if not (or (eq $key "user_id") (eq $key "description") (eq $key "title"))}} *{{ $key }}:* {{ $value }}\n {{- end -}} {{- end -}}"
    }
  }
]
