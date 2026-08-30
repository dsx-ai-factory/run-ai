#!/bin/bash
echo "Building OpenAPI services after updating the OpenAPI template"
find . -name swagger | grep -v -E "./infra/|./redoc|./frontend|./backend/js/" | xargs dirname | xargs -I {} bash -c 'cd {} && make openapi-build'

echo
echo "OpenAPI services built successfully"
