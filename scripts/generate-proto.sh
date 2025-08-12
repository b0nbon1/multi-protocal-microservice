#!/bin/bash

# Script to generate gRPC Go code from proto files

set -e

echo "🔧 Installing dependencies..."

# Install dependencies for NestJS services
for service in auth contract payment dispute; do
    echo "Installing dependencies for $service service..."
    cd services/$service && npm install && cd ../..
done

# Create the output directory for Go gateway
mkdir -p gateway/internal/grpc/{auth,contract,payment,dispute,notification,audit}

echo "🔄 Generating Go gRPC code..."

# Generate Go code for gateway
protoc --go_out=gateway/internal/grpc/auth --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/auth --go-grpc_opt=paths=source_relative \
       proto/auth.proto

protoc --go_out=gateway/internal/grpc/contract --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/contract --go-grpc_opt=paths=source_relative \
       proto/contract.proto

protoc --go_out=gateway/internal/grpc/payment --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/payment --go-grpc_opt=paths=source_relative \
       proto/payment.proto

protoc --go_out=gateway/internal/grpc/dispute --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/dispute --go-grpc_opt=paths=source_relative \
       proto/dispute.proto

protoc --go_out=gateway/internal/grpc/notification --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/notification --go-grpc_opt=paths=source_relative \
       proto/notification.proto

protoc --go_out=gateway/internal/grpc/audit --go_opt=paths=source_relative \
       --go-grpc_out=gateway/internal/grpc/audit --go-grpc_opt=paths=source_relative \
       proto/audit.proto

echo "✅ gRPC Go code generated successfully!"

# Create output directories for NestJS services
for service in auth contract payment dispute; do
    mkdir -p services/$service/src/proto
done

# Generate TypeScript/JavaScript code for NestJS services
echo "🔄 Generating TypeScript code for NestJS services..."

for service in auth contract payment dispute; do
    echo "Generating for $service service..."
    
    # Use the local ts-proto installation
    protoc --plugin=protoc-gen-ts_proto=services/$service/node_modules/.bin/protoc-gen-ts_proto \
           --ts_proto_out=services/$service/src/proto \
           --ts_proto_opt=nestJs=true \
           --ts_proto_opt=addGrpcMetadata=true \
           proto/$service.proto || echo "⚠️ Failed to generate for $service"
done

echo "✅ All protobuf files generated!"
