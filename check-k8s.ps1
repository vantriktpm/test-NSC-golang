# Script kiểm tra trạng thái Kubernetes cluster

Write-Host "🔍 Kiểm tra trạng thái Kubernetes cluster..." -ForegroundColor Green

# Kiểm tra kubectl
try {
    $kubectlVersion = kubectl version --client --output=json | ConvertFrom-Json
    Write-Host "✅ kubectl version: $($kubectlVersion.clientVersion.gitVersion)" -ForegroundColor Green
} catch {
    Write-Host "❌ kubectl không được cài đặt hoặc không hoạt động" -ForegroundColor Red
    exit 1
}

# Kiểm tra kết nối cluster
try {
    kubectl cluster-info | Out-Null
    Write-Host "✅ Cluster connection: OK" -ForegroundColor Green
} catch {
    Write-Host "❌ Không thể kết nối đến cluster" -ForegroundColor Red
    Write-Host "💡 Gợi ý: Khởi động Docker Desktop hoặc minikube" -ForegroundColor Yellow
    exit 1
}

# Kiểm tra nodes
Write-Host ""
Write-Host "📊 Cluster nodes:" -ForegroundColor Cyan
kubectl get nodes

# Kiểm tra namespaces
Write-Host ""
Write-Host "📦 Namespaces:" -ForegroundColor Cyan
kubectl get namespaces

# Kiểm tra pods trong namespace url-shortener (nếu có)
Write-Host ""
Write-Host "🔍 Kiểm tra namespace url-shortener:" -ForegroundColor Cyan
try {
    kubectl get pods -n url-shortener
    kubectl get services -n url-shortener
    kubectl get ingress -n url-shortener
} catch {
    Write-Host "ℹ️ Namespace url-shortener chưa tồn tại" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "✅ Cluster đã sẵn sàng để triển khai!" -ForegroundColor Green
