# Script PowerShell để triển khai URL Shortener lên Kubernetes

Write-Host "🚀 Bắt đầu triển khai URL Shortener lên Kubernetes..." -ForegroundColor Green

# Kiểm tra kubectl
try {
    kubectl version --client | Out-Null
    Write-Host "✅ kubectl đã được cài đặt" -ForegroundColor Green
} catch {
    Write-Host "❌ kubectl không được tìm thấy. Vui lòng cài đặt kubectl trước." -ForegroundColor Red
    exit 1
}

# Kiểm tra kết nối cluster
try {
    kubectl cluster-info | Out-Null
    Write-Host "✅ Kubernetes cluster đã sẵn sàng" -ForegroundColor Green
} catch {
    Write-Host "❌ Không thể kết nối đến Kubernetes cluster. Vui lòng kiểm tra kubeconfig." -ForegroundColor Red
    Write-Host "💡 Gợi ý: Khởi động Docker Desktop hoặc minikube" -ForegroundColor Yellow
    exit 1
}

# Tạo namespace
Write-Host "📦 Tạo namespace..." -ForegroundColor Blue
kubectl apply -f k8s/namespace.yaml

# Triển khai PostgreSQL
Write-Host "🐘 Triển khai PostgreSQL..." -ForegroundColor Blue
kubectl apply -f k8s/postgres-deployment.yaml

# Triển khai Redis
Write-Host "🔴 Triển khai Redis..." -ForegroundColor Blue
kubectl apply -f k8s/redis-deployment.yaml

# Chờ database sẵn sàng
Write-Host "⏳ Chờ database sẵn sàng..." -ForegroundColor Yellow
kubectl wait --for=condition=ready pod -l app=postgres -n url-shortener --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n url-shortener --timeout=300s

# Tạo secrets và configmap
Write-Host "🔐 Tạo secrets và configmap..." -ForegroundColor Blue
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/configmap.yaml

# Triển khai ứng dụng chính
Write-Host "🌐 Triển khai ứng dụng chính..." -ForegroundColor Blue
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Chờ ứng dụng sẵn sàng
Write-Host "⏳ Chờ ứng dụng sẵn sàng..." -ForegroundColor Yellow
kubectl wait --for=condition=ready pod -l app=url-shortener -n url-shortener --timeout=300s

# Tạo Ingress (tùy chọn)
Write-Host "🌍 Tạo Ingress..." -ForegroundColor Blue
kubectl apply -f k8s/ingress.yaml

Write-Host "✅ Triển khai hoàn tất!" -ForegroundColor Green
Write-Host ""

Write-Host "📊 Trạng thái pods:" -ForegroundColor Cyan
kubectl get pods -n url-shortener

Write-Host ""
Write-Host "🔗 Services:" -ForegroundColor Cyan
kubectl get services -n url-shortener

Write-Host ""
Write-Host "🌐 Ingress:" -ForegroundColor Cyan
kubectl get ingress -n url-shortener

Write-Host ""
Write-Host "📝 Để truy cập ứng dụng:" -ForegroundColor Yellow
Write-Host "1. Port forward: kubectl port-forward -n url-shortener service/url-shortener-service 8080:80" -ForegroundColor White
Write-Host "2. Truy cập: http://localhost:8080" -ForegroundColor White
Write-Host "3. Health check: http://localhost:8080/api/v1/health" -ForegroundColor White

Write-Host ""
Write-Host "🧪 Test ứng dụng:" -ForegroundColor Yellow
Write-Host "curl -X POST http://localhost:8080/api/v1/shorten -H 'Content-Type: application/json' -d '{\"url\": \"https://example.com\"}'" -ForegroundColor White
