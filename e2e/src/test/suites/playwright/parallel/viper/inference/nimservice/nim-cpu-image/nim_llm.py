#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


import json
import time
import random
import math
import threading
import asyncio
from datetime import datetime
from typing import Dict, Any, List, Optional
from concurrent import futures
from contextlib import contextmanager, asynccontextmanager

from fastapi import FastAPI, HTTPException, Response
from pydantic import BaseModel
import uvicorn
from prometheus_client import Counter, Gauge, Histogram, generate_latest, CONTENT_TYPE_LATEST

# gRPC imports
import grpc
import grpc.aio
from grpc_reflection.v1alpha import reflection
import os
import sys

# Constants
MODEL_NAME = "meta/llama-3.2-1b-instruct"
HTTP_PORT = 8000
GRPC_PORT = 8001
NIM_REQUEST_LATENCY = float(os.getenv("NIM_REQUEST_LATENCY", "2.0"))

# Import generated gRPC modules (generated at build time)
import nim_service_pb2
import nim_service_pb2_grpc

# Initialize FastAPI app
app = FastAPI(title="NVIDIA NIM", description="Simulated NVIDIA NIM API", version="1.0.0")

# Prometheus metrics
num_requests_running = Gauge('num_requests_running', 'Number of requests currently running')
num_requests_waiting = Gauge('num_requests_waiting', 'Number of requests waiting in queue')
num_request_max = Gauge('num_request_max', 'Maximum number of concurrent requests supported')
request_success_total = Counter('request_success_total', 'Total number of successful requests')
request_failure_total = Counter('request_failure_total', 'Total number of failed requests')
gpu_cache_usage_perc = Gauge('gpu_cache_usage_perc', 'GPU cache usage percentage')
time_to_first_token_seconds = Histogram('time_to_first_token_seconds', 'Time to first token in seconds', 
                                       buckets=[0.05, 0.1, 0.15, 0.2, 0.3, 0.5, 1.0, 2.0, 5.0])
e2e_request_latency_seconds = Histogram('e2e_request_latency_seconds', 'End-to-end request latency in seconds',
                                       buckets=[0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0])

# Initialize static/semi-static metrics
num_request_max.set(100)  # Maximum concurrent requests supported

# Internal state for realistic metric simulation
_start_time = time.time()
_total_requests_made = 0

# Request/Response models for chat completions
class ChatMessage(BaseModel):
    role: str
    content: str

class ChatCompletionRequest(BaseModel):
    model: str
    messages: List[ChatMessage]
    max_tokens: Optional[int] = None
    temperature: Optional[float] = None
    stream: Optional[bool] = False

class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: List[Dict[str, Any]]
    usage: Dict[str, int]

@contextmanager
def track_grpc_request():
    """Track gRPC request with simulated latency"""
    num_requests_running.inc()
    try:
        time.sleep(NIM_REQUEST_LATENCY)
        yield
    finally:
        num_requests_running.dec()

@asynccontextmanager
async def track_http_request():
    """Track HTTP request with simulated latency and concurrency support"""
    num_requests_running.inc()
    try:
        await asyncio.sleep(NIM_REQUEST_LATENCY)
        yield
    finally:
        num_requests_running.dec()

def update_realistic_metrics():
    """Update metrics with realistic simulated values"""
    global _total_requests_made
    
    # Simulate running and waiting requests based on some activity pattern
    uptime_minutes = (time.time() - _start_time) / 60
    base_activity = max(0, 0.5 + 0.3 * (1 + math.sin(uptime_minutes * 0.1)))  # Sine wave activity
    
    # Requests waiting (usually 0-3, can spike higher during busy times)
    waiting = max(0, int((base_activity - 0.7) * 5 + random.gauss(0, 0.5))) if base_activity > 0.7 else 0
    waiting = min(waiting, 15)
    num_requests_waiting.set(waiting)
    
    # GPU cache usage (70-95%, with some variance)
    cache_usage = 75 + 15 * base_activity + random.gauss(0, 3)
    cache_usage = max(60, min(95, cache_usage))
    gpu_cache_usage_perc.set(cache_usage)

def simulate_request_metrics(success: bool = True):
    """Simulate request completion with realistic timing metrics"""
    global _total_requests_made
    _total_requests_made += 1
    
    if success:
        request_success_total.inc()
        
        # Time to first token: ~100ms average, 10ms std dev
        ttft = max(0.05, random.gauss(0.1, 0.01))
        time_to_first_token_seconds.observe(ttft)
        
        # E2E latency: depends on response length, ~1s average with variance
        e2e_latency = max(0.2, random.gauss(1.0, 0.3))
        e2e_request_latency_seconds.observe(e2e_latency)
    else:
        request_failure_total.inc()
        # Failed requests typically fail fast
        time_to_first_token_seconds.observe(random.uniform(0.01, 0.05))
        e2e_request_latency_seconds.observe(random.uniform(0.05, 0.2))

# gRPC Service Implementation
class NimServiceServicer(nim_service_pb2_grpc.NimServiceServicer):
    """gRPC service implementation for NIM service"""
    
    def ListModels(self, request, context):
        """List available models"""
        update_realistic_metrics()
        simulate_request_metrics(success=True)
        
        model = nim_service_pb2.ModelInfo(
            id=MODEL_NAME,
            name=MODEL_NAME,
            created=int(time.time()),
            owned_by="nim-service"
        )
        
        return nim_service_pb2.ListModelsResponse(models=[model])
    
    def ChatCompletion(self, request, context):
        """Chat completion endpoint"""
        update_realistic_metrics()
        
        with track_grpc_request():
            # Simulate occasional failures (2% failure rate)
            if random.random() < 0.02:
                simulate_request_metrics(success=False)
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Internal server error - model temporarily unavailable")
                return nim_service_pb2.ChatCompletionResponse()
            
            # Generate dummy response
            dummy_response = "This is a simulated gRPC response from the NVIDIA NIM server. The model is running successfully."
            response_id = f"chatcmpl-grpc-{int(time.time())}{random.randint(1000, 9999)}"
            
            simulate_request_metrics(success=True)
            
            return nim_service_pb2.ChatCompletionResponse(
                id=response_id,
                model=request.model if request.model else MODEL_NAME,
                created=int(time.time()),
                content=dummy_response,
                prompt_tokens=50,
                completion_tokens=25,
                total_tokens=75
            )
    
    def Inference(self, request, context):
        """Simple text inference"""
        update_realistic_metrics()
        
        with track_grpc_request():
            # Simulate occasional failures (2% failure rate)
            if random.random() < 0.02:
                simulate_request_metrics(success=False)
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Internal server error - model temporarily unavailable")
                return nim_service_pb2.InferenceResponse()
            
            # Generate dummy response based on prompt
            response_text = f"Processed prompt: '{request.prompt}'. This is a simulated inference response from NIM gRPC service."
            response_id = f"inf-{int(time.time())}{random.randint(1000, 9999)}"
            
            simulate_request_metrics(success=True)
            
            return nim_service_pb2.InferenceResponse(
                id=response_id,
                model=request.model if request.model else MODEL_NAME,
                created=int(time.time()),
                response_text=response_text,
                tokens_used=len(request.prompt.split()) + 15  # Simulate token usage
            )
    
    def Health(self, request, context):
        """Health check"""
        update_realistic_metrics()
        
        return nim_service_pb2.HealthResponse(
            status="healthy",
            model=MODEL_NAME,
            timestamp=int(time.time())
        )

@app.get("/v1/metrics")
async def metrics():
    """Prometheus metrics endpoint"""
    # Update metrics with current realistic values
    update_realistic_metrics()
    
    # Generate and return Prometheus metrics format
    metrics_output = generate_latest()
    return Response(content=metrics_output, media_type=CONTENT_TYPE_LATEST)

@app.get("/v1/models")
async def list_models():
    """List available models - OpenAI compatible endpoint"""
    update_realistic_metrics()
    simulate_request_metrics(success=True)
    
    return {
        "object": "list",
        "data": [
            {
                "id": MODEL_NAME,
                "object": "model",
                "created": int(time.time()),
                "owned_by": "nim-service",
                "permission": [],
                "root": MODEL_NAME,
                "parent": None
            }
        ]
    }

@app.post("/v1/chat/completions")
async def chat_completions(request: ChatCompletionRequest):
    """Chat completions endpoint - OpenAI compatible"""
    update_realistic_metrics()
    
    async with track_http_request():
        # Simulate occasional failures (2% failure rate)
        if random.random() < 0.02:
            simulate_request_metrics(success=False)
            raise HTTPException(status_code=500, detail="Internal server error - model temporarily unavailable")
        
        # Dummy response content
        dummy_response_content = "This is a simulated response from the NVIDIA NIM server. The model is running successfully and ready to handle requests."
        
        # Generate a dummy response ID
        response_id = f"chatcmpl-{int(time.time())}{random.randint(1000, 9999)}"
        
        response = {
            "id": response_id,
            "object": "chat.completion",
            "created": int(time.time()),
            "model": request.model if request.model else MODEL_NAME,
            "choices": [
                {
                    "index": 0,
                    "message": {
                        "role": "assistant",
                        "content": dummy_response_content
                    },
                    "finish_reason": "stop"
                }
            ],
            "usage": {
                "prompt_tokens": 50,  # dummy values
                "completion_tokens": 25,
                "total_tokens": 75
            }
        }
        
        # Track successful request
        simulate_request_metrics(success=True)
        return response

@app.get("/v1/health/ready")
async def health_ready():
    """Health check endpoint for readiness"""
    update_realistic_metrics()
    return {
        "status": "ready",
        "model": MODEL_NAME,
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }

@app.get("/v1/health/live")
async def health_live():
    """Health check endpoint for liveness"""
    return {
        "status": "alive",
        "model": MODEL_NAME,
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }

@app.get("/health")
async def health():
    """Alternative health check endpoint"""
    update_realistic_metrics()
    return {"status": "healthy"}

@app.get("/")
async def root():
    """Root endpoint"""
    update_realistic_metrics()
    return {
        "message": "NVIDIA NIM is running",
        "model": MODEL_NAME,
        "endpoints": ["/v1/models", "/v1/chat/completions", "/v1/health/ready", "/v1/health/live", "/v1/metrics"],
        "grpc_port": GRPC_PORT,
        "grpc_services": ["ListModels", "ChatCompletion", "Inference", "Health"]
    }

def start_grpc_server():
    """Start gRPC server"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Add our service implementation to the server
    nim_service_pb2_grpc.add_NimServiceServicer_to_server(NimServiceServicer(), server)
    
    # Get service name and enable reflection
    service_descriptor = nim_service_pb2.DESCRIPTOR.services_by_name['NimService']
    service_full_name = service_descriptor.full_name
    
    # Enable basic gRPC reflection for service discovery
    SERVICE_NAMES = (
        service_full_name,         # Our service: nim_service.NimService
        reflection.SERVICE_NAME,   # Reflection service itself
    )
    
    reflection.enable_server_reflection(SERVICE_NAMES, server)
    print(f"gRPC reflection enabled for: {service_full_name}")
    
    listen_addr = f'[::]:{GRPC_PORT}'
    server.add_insecure_port(listen_addr)
    server.start()
    print(f"gRPC server started on port {GRPC_PORT}")
    
    # Start server in background thread
    def serve():
        try:
            server.wait_for_termination()
        except KeyboardInterrupt:
            server.stop(grace=5)
    
    thread = threading.Thread(target=serve, daemon=True)
    thread.start()
    
    return server

def start_http_server():
    """Start HTTP server"""
    print(f"HTTP server starting on port {HTTP_PORT}")
    uvicorn.run(app, host="0.0.0.0", port=HTTP_PORT)

if __name__ == "__main__":
    print("Starting NVIDIA NIM simulation servers...")
    print(f"HTTP server will be available on port {HTTP_PORT}")
    print(f"gRPC server will be available on port {GRPC_PORT}")
    
    # Start gRPC server in background
    grpc_server = start_grpc_server()
    
    # Give gRPC server time to start
    time.sleep(1)
    
    # Start HTTP server in main thread (blocking)
    try:
        start_http_server()
    except KeyboardInterrupt:
        print("\nShutting down servers...")
        grpc_server.stop(grace=5)
    finally:
        grpc_server.stop(grace=5)
