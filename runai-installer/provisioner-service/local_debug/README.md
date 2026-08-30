To run the provisioning service locally, you'll need to:
- Create a folder for the secrets, i.e: `/tmp/runai-certs`
- Ensure you have the following files:
    - `tls.crt`
    - `tls.key`
    - `wildcard.crt`
    - `wildcard.key`
    - `runai_jfrog_token`

- In the `launch.json` file, add the following:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Provisioner Service",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/provisioner-service/cmd",
            "env": {
                "PORT": "8080",
                "SECRETS_MOUNT_PATH": "/tmp/runai-certs",
                "RUNAI_TLS_CERT_PATH": "/tmp/runai-certs/tls.crt",
                "RUNAI_TLS_KEY_PATH": "/tmp/runai-certs/tls.key",
                "RUNAI_WILDCARD_TLS_CERT_PATH": "/tmp/runai-certs/wildcard.crt",
                "RUNAI_WILDCARD_TLS_KEY_PATH": "/tmp/runai-certs/wildcard.key",
                "RUNAI_JFROG_TOKEN_PATH": "/tmp/runai-certs/jfrog-token"
            },
            "args": [],
            "showLog": true,
        }
    ]
}
```

That's it, your service should now run with debug mode.