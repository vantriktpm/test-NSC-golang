#!/bin/bash

# Script để triển khai URL Shortener lên Kubernetes

echo "🚀 Bắt đầu triển khai URL Shortener lên Kubernetes..."

# Kiểm tra kubectl
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl không được tìm thấy. Vui lòng cài đặt kubectl trước."
    exit 1
fi

# Kiểm tra kết nối cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Không thể kết nối đến Kubernetes cluster. Vui lòng kiểm tra kubeconfig."
    exit 1
fi

echo "✅ Kubernetes cluster đã sẵn sàng"

# Tạo namespace
echo "📦 Tạo namespace..."
kubectl apply -f k8s/namespace.yaml

# Triển khai PostgreSQL
echo "🐘 Triển khai PostgreSQL..."
kubectl apply -f k8s/postgres-deployment.yaml

# Triển khai Redis
echo "🔴 Triển khai Redis..."
kubectl apply -f k8s/redis-deployment.yaml

# Chờ database sẵn sàng
echo "⏳ Chờ database sẵn sàng..."
kubectl wait --for=condition=ready pod -l app=postgres -n url-shortener --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n url-shortener --timeout=300s

# Tạo secrets và configmap
echo "🔐 Tạo secrets và configmap..."
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/configmap.yaml

# Triển khai ứng dụng chính
echo "🌐 Triển khai ứng dụng chính..."
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Chờ ứng dụng sẵn sàng
echo "⏳ Chờ ứng dụng sẵn sàng..."
kubectl wait --for=condition=ready pod -l app=url-shortener -n url-shortener --timeout=300s

# Tạo Ingress (tùy chọn)
echo "🌍 Tạo Ingress..."
kubectl apply -f k8s/ingress.yaml

echo "✅ Triển khai hoàn tất!"
echo ""
echo "📊 Trạng thái pods:"
kubectl get pods -n url-shortener

echo ""
echo "🔗 Services:"
kubectl get services -n url-shortener

echo ""
echo "🌐 Ingress:"
kubectl get ingress -n url-shortener

echo ""
echo "📝 Để truy cập ứng dụng:"
echo "1. Port forward: kubectl port-forward -n url-shortener service/url-shortener-service 8080:80"
echo "2. Truy cập: http://localhost:8080"
echo "3. Health check: http://localhost:8080/api/v1/health"
