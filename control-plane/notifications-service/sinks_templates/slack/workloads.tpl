[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": {{ printf "%q" (printf "*%s*" (or .EventProperties.title "")) }}
    }
  }
  {{- with .EventProperties.description -}}
  ,
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": {{ printf "%q" . }}
    }
  }
  {{- end }},
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": {{ printf "%q" (printf "*Name:* %s" (or .EventProperties.name "")) }}
      },
      {
        "type": "mrkdwn",
        "text": {{ printf "%q" (printf "*Type:* %s" (or .EventProperties.workload_type "")) }}
      },
      {
        "type": "mrkdwn",
        "text": {{ printf "%q" (printf "*Project:* %s" (or .EventProperties.project "")) }}
      },
      {
        "type": "mrkdwn",
        "text": {{ printf "%q" (printf "*Created by:* %s" (or .EventProperties.user_id "")) }}
      },
      {
        "type": "mrkdwn",
        "text": {{ printf "%q" (printf "*Time:* %s" (or .EventProperties.time "")) }}
      }
    ]
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": {{ printf "%q" (printf "For more details, visit: <%s|Workload Details>" (or .EventProperties.workload_url "")) }}
    }
  }
]
