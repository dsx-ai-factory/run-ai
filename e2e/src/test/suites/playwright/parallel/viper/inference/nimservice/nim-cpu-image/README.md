# NVIDIA NIM Simulation

This is a simulated NVIDIA NIM image that mimics the behavior of an OpenAI-compatible API server with both HTTP and gRPC interfaces.

## Features

- **Model**: `meta/llama-3.2-1b-instruct`
- **HTTP Port**: 8000
- **gRPC Port**: 8001
- **HTTP Endpoints**:
  - `GET /v1/models` - List available models
  - `POST /v1/chat/completions` - Chat completions with dummy responses
  - `GET /v1/health/ready` - Readiness health check endpoint
  - `GET /v1/health/live` - Liveness health check endpoint
  - `GET /v1/metrics` - Prometheus metrics endpoint
- **gRPC Services**:
  - `ListModels()` - List available models
  - `ChatCompletion()` - Chat completions (similar to HTTP API)
  - `Inference()` - Simple text inference
  - `Health()` - Health check
  - **Service Discovery**: gRPC reflection enabled for listing services and methods

## Building and Running

### Build the Docker image:
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t nim-cpu .
```

### Run the container:
```bash
docker run -p 8000:8000 -p 8001:8001 -e NIM_REQUEST_LATENCY=2.0 nim-cpu
```

**Environment Variables:**
- `NIM_REQUEST_LATENCY` (float, default: 2.0): Request processing latency in seconds. Controls how long each request "hangs" to simulate processing time.

### Test the endpoints:

1. List models:
```bash
curl http://localhost:8000/v1/models
```

2. Chat completion:
```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta/llama-3.2-1b-instruct",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

3. Health check (readiness):
```bash
curl http://localhost:8000/v1/health/ready
```

4. Health check (liveness):
```bash
curl http://localhost:8000/v1/health/live
```

5. Metrics (Prometheus format):
```bash
curl http://localhost:8000/v1/metrics
```

### Test gRPC endpoints:

**Prerequisites**: Install gRPC tools:
```bash
pip install grpcio-tools grpcurl
```

1. List gRPC services (service discovery):
```bash
grpcurl -plaintext localhost:8001 list
```

2. List gRPC methods for NimService:
```bash
grpcurl -plaintext localhost:8001 list nim_service.NimService
```

3. List models via gRPC:
```bash
grpcurl -plaintext localhost:8001 nim_service.NimService/ListModels
```

4. Simple inference via gRPC:
```bash
grpcurl -plaintext -d '{"model": "meta/llama-3.2-1b-instruct", "prompt": "Hello, world!"}' \
  localhost:8001 nim_service.NimService/Inference
```

5. Chat completion via gRPC:
```bash
grpcurl -plaintext -d '{
  "model": "meta/llama-3.2-1b-instruct",
  "messages": [{"role": "user", "content": "Hello!"}]
}' localhost:8001 nim_service.NimService/ChatCompletion
```

6. Health check via gRPC:
```bash
grpcurl -plaintext localhost:8001 nim_service.NimService/Health
```

## Metrics

The server exposes Prometheus-compatible metrics at `/v1/metrics`:

- **`num_requests_running`** (Gauge) - The exact number of concurrent requests currently being processed by the server (both HTTP and gRPC). This metric scales directly with the number of incoming requests and their duration (controlled by the environment variable `NIM_REQUEST_LATENCY`, in seconds).
- **`num_requests_waiting`** (Gauge) - Requests waiting in queue (Simulated)
- **`num_request_max`** (Gauge) - Maximum concurrent requests supported
- **`request_success_total`** (Counter) - Total successful requests
- **`request_failure_total`** (Counter) - Total failed requests
- **`gpu_cache_usage_perc`** (Gauge) - GPU cache usage percentage (Simulated)
- **`time_to_first_token_seconds`** (Histogram) - Time to first token latency
- **`e2e_request_latency_seconds`** (Histogram) - End-to-end request latency

Most metrics are simulated, but `num_requests_running` is real:

- **`num_requests_running`**: Real-time counter of active concurrent requests.
- **Other metrics**: Based on sine wave activity cycles with Gaussian noise for realism.
- **Request timing**: TTFT ~100ms ±10ms, E2E latency ~1s with variance  
- **Failure simulation**: 2% random failure rate with faster failure response times
- **GPU usage**: Varies 75-95% based on simulated load with natural fluctuations

## Development

To run locally without Docker:
```bash
pip install -r requirements.txt
python nim_llm.py
```

The server will start both HTTP (port 8000) and gRPC (port 8001) interfaces simultaneously.

## gRPC Protocol

The gRPC service is defined in `nim_service.proto` and includes:

- **Service Discovery**: gRPC reflection is enabled, allowing clients to list available services and methods
- **ListModels**: Returns available model information
- **ChatCompletion**: OpenAI-style chat completion with message history
- **Inference**: Simple text-to-text inference for basic prompts
- **Health**: Service health check

**Note**: gRPC Python code (`nim_service_pb2.py` and `nim_service_pb2_grpc.py`) is generated automatically during Docker build time for better reliability.

All gRPC methods include the same realistic metrics simulation and 2% failure rate as the HTTP endpoints.
