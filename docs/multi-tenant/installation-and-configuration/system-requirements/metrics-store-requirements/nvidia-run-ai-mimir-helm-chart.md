---
hidden: true
---

# NVIDIA Run:ai Mimir Helm Chart

```yaml
metaMonitoring:
 serviceMonitor:
   clusterLabel: "<<REPLACE ME WITH CLUSTER IDENTIFIER>>" #example: “cluster-prod01”
   enabled: true

alertmanager:
 enabled: false

minio:
 enabled: false

query_scheduler:
 replicas: 2
 resources:
   requests:
     cpu: 1
     memory: 1Gi
 enabled: true

querier:
 replicas: 3
 resources:
   limits:
     memory: 8.5Gi
   requests:
     cpu: 2
     memory: 6Gi

ruler:
 enabled: false

query_frontend:
 replicas: 3
 resources:
   limits:
     memory: 2.8Gi
   requests:
     cpu: 2
     memory: 2Gi

ingester:
 persistentVolume:
   size: 100Gi
 replicas: 3
 resources:
   limits:
     memory: 12Gi
   requests:
     cpu: 3.5
     memory: 8Gi

admin-cache:
 enabled: true
 replicas: 2

chunks-cache:
 enabled: true
 replicas: 4

index-cache:
 enabled: true
 replicas: 4

metadata-cache:
 enabled: true

results-cache:
 enabled: true
 replicas: 4
 allocatedMemory: 1024

overrides_exporter:
 replicas: 1
 resources:
   limits:
     memory: 128Mi
   requests:
     cpu: 100m
     memory: 128Mi

compactor:
 resources:
   limits:
     memory: 2Gi
   requests:
     cpu: 1
     memory: 1Gi
 persistentVolume:
   size: 300Gi

store_gateway:
 persistentVolume:
   size: 100Gi
 replicas: 3
 resources:
   limits:
     memory: 8.5Gi
   requests:
     cpu: 1
     memory: 6Gi

distributor:
 replicas: 3
 resources:
   limits:
     memory: 5.7Gi
   requests:
     cpu: 2
     memory: 4Gi

nginx:
 enabled: true
 replicas: 3
 verboseLogging: false
 resources:
   requests:
     cpu: 1000m
     memory: 300Mi
 ingress:
   enabled: true
   ingressClassName: nginx
   annotations:
     nginx.ingress.kubernetes.io/client-body-buffer-size: 128k
   hosts:
    - host: <MIMIR_HOST_FQDN> #example: mimir.example.com
      paths:
        - path: /
          pathType: Prefix
   tls:
    - secretName: mimir-tls-certificate
      hosts:
        - <MIMIR_HOST_FQDN> #example: mimir.example.com

 nginxConfig:
   logFormat: |-
     main escape=json
     '{'
       '"time_local":"$time_local",'
       '"remote_addr":"$remote_addr",'
       '"remote_user":"$remote_user",'
       '"request":"$request",'
       '"status": "$status",'
       '"body_bytes_sent":"$body_bytes_sent",'
       '"request_time":"$request_time",'
       '"http_referrer":"$http_referer",'
       '"http_user_agent":"$http_user_agent"'
     '}';
   file: |
     worker_processes  5;  ## Default: 1
     error_log  /dev/stderr;
     pid        /tmp/nginx.pid;
     worker_rlimit_nofile 8192;

     events {
       worker_connections  4096;  ## Default: 1024
     }

     http {
       client_body_temp_path /tmp/client_temp;
       proxy_temp_path       /tmp/proxy_temp_path;
       fastcgi_temp_path     /tmp/fastcgi_temp;
       uwsgi_temp_path       /tmp/uwsgi_temp;
       scgi_temp_path        /tmp/scgi_temp;

       default_type application/octet-stream;
       log_format   {{ .Values.nginx.nginxConfig.logFormat }}

       {{- if .Values.nginx.verboseLogging }}
       access_log   /dev/stderr  main;
       {{- else }}

       map $status $loggable {
         ~^[23]  0;
         default 1;
       }
       access_log   /dev/stderr  main  if=$loggable;
       {{- end }}

       sendfile     on;
       tcp_nopush   on;
       resolver {{ .Values.global.dnsService }}.{{ .Values.global.dnsNamespace }}.svc.{{ .Values.global.clusterDomain }};

       {{- with .Values.nginx.nginxConfig.httpSnippet }}
       {{ . | nindent 2 }}
       {{- end }}

       server {
         listen 8080;

         {{- if .Values.nginx.basicAuth.enabled }}
         auth_basic           "Mimir";
         auth_basic_user_file /etc/nginx/secrets/.htpasswd;
         {{- end }}

         location = / {
           return 200 'OK';
           auth_basic off;
         }

         {{- if not (include "mimir.calculatedConfig" . | fromYaml).multitenancy_enabled }}
         proxy_set_header X-Scope-OrgID 0;
         {{- end }}

         # Distributor endpoints
         location /distributor {
           proxy_pass      http://{{ template "mimir.fullname" . }}-distributor-headless.{{ .Release.Namespace }}.svc.{{ .Values.global.clusterDomain }}:{{ include "mimir.serverHttpListenPort" . }}$request_uri;
         }
         location = /api/v1/push {
           proxy_pass      http://{{ template "mimir.fullname" . }}-distributor-headless.{{ .Release.Namespace }}.svc.{{ .Values.global.clusterDomain }}:{{ include "mimir.serverHttpListenPort" . }}$request_uri;
         }
       }
     }

mimir:
 structuredConfig:
   multitenancy_enabled: true

   limits:
     compactor_block_upload_enabled: true
     max_global_series_per_user: 3000000
     ingestion_rate: 200000
     ingestion_burst_size: 600000
     max_query_parallelism: 240
     max_fetched_chunks_per_query: 0

   activity_tracker:
     filepath: /data/metrics-activity.log

   compactor:
     data_dir: /data

   ingester:
     ring:
       zone_awareness_enabled: true

   runtime_config:
     file: /var/mimir/runtime.yaml

   blocks_storage:
     backend: s3
     s3:
       bucket_name: <<REPLACE ME WITH S3 BUCKET NAME>> #example: my-bucket
       endpoint: s3.dualstack.us-east-1.amazonaws.com
     tsdb:
       dir: /data/tsdb/
     bucket_store:
       sync_dir: /data/tsdb-sync/

   query_scheduler:
     max_outstanding_requests_per_tenant: 4000

   frontend:
     parallelize_shardable_queries: true

   server:
     log_format: json

   store_gateway:
     sharding_ring:
       tokens_file_path: /data/tokens
       zone_awareness_enabled: true
```
