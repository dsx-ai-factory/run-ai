# Use Python slim base image
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements first for better Docker layer caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the proto file and generate gRPC code at build time
COPY nim_service.proto .

# Generate gRPC code during Docker build (more reliable)
RUN python -m grpc_tools.protoc \
    --proto_path=. \
    --python_out=. \
    --grpc_python_out=. \
    nim_service.proto

# Verify files were generated
RUN ls -la nim_service_pb2* && echo "gRPC files generated successfully"

# Copy the application code
COPY nim_llm.py .

# Expose the ports
EXPOSE 8000 8001

# Set environment variables
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
  CMD python -c "import requests; requests.get('http://localhost:8000/v1/health/ready')" || exit 1

# Run the application
CMD ["python", "nim_llm.py"]
