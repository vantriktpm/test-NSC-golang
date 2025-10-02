#!/bin/bash

# Script khởi động ngrok tunnel cho GitHub Actions testing

PORT=${1:-8080}
SERVICE=${2:-url-shortener-service}
NAMESPACE=${3:-url-shortener}

echo "🚀 Starting ngrok tunnel for GitHub Actions..."

# Check if ngrok is installed
if ! command -v ngrok &> /dev/null; then
    echo "❌ ngrok not found. Please install ngrok first."
    echo "💡 Install with: brew install ngrok (Mac) or download from https://ngrok.com/download"
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Check if Kubernetes cluster is running
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Kubernetes cluster not accessible"
    echo "💡 Make sure Docker Desktop Kubernetes is enabled"
    exit 1
fi

# Check if service exists
if ! kubectl get service $SERVICE -n $NAMESPACE &> /dev/null; then
    echo "❌ Service $SERVICE not found in namespace $NAMESPACE"
    echo "💡 Available services:"
    kubectl get services -n $NAMESPACE
    exit 1
fi

echo ""
echo "📋 Configuration:"
echo "  Port: $PORT"
echo "  Service: $SERVICE"
echo "  Namespace: $NAMESPACE"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    if [ ! -z "$PORT_FORWARD_PID" ]; then
        kill $PORT_FORWARD_PID 2>/dev/null
    fi
    echo "✅ Cleanup completed"
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM EXIT

# Start port forward in background
echo "📡 Starting port forward..."
kubectl port-forward -n $NAMESPACE service/$SERVICE $PORT:80 &
PORT_FORWARD_PID=$!

# Wait for port forward to start
echo "⏳ Waiting for port forward to start..."
sleep 5

# Check if port forward is working
if curl -s -f "http://localhost:$PORT/api/v1/health" > /dev/null; then
    echo "✅ Port forward is working"
else
    echo "⚠️ Port forward might not be ready yet, continuing..."
fi

echo ""
echo "🌐 Starting ngrok tunnel..."
echo "💡 Copy the HTTPS URL from ngrok output for GitHub Actions"
echo "💡 Press Ctrl+C to stop the tunnel"
echo ""

# Start ngrok
ngrok http $PORT
