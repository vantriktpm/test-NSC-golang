# Script thiết lập Kubernetes cluster local cho Windows

Write-Host "🚀 Thiết lập Kubernetes cluster local..." -ForegroundColor Green

# Kiểm tra Docker Desktop
Write-Host "🔍 Kiểm tra Docker Desktop..." -ForegroundColor Blue
try {
    docker version | Out-Null
    Write-Host "✅ Docker đã được cài đặt" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker không được cài đặt hoặc không chạy" -ForegroundColor Red
    Write-Host "💡 Vui lòng cài đặt Docker Desktop và khởi động lại" -ForegroundColor Yellow
    exit 1
}

# Kiểm tra kubectl
Write-Host "🔍 Kiểm tra kubectl..." -ForegroundColor Blue
try {
    kubectl version --client | Out-Null
    Write-Host "✅ kubectl đã được cài đặt" -ForegroundColor Green
} catch {
    Write-Host "❌ kubectl không được cài đặt" -ForegroundColor Red
    Write-Host "💡 Cài đặt kubectl từ: https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/" -ForegroundColor Yellow
    exit 1
}

# Kiểm tra Docker Desktop Kubernetes
Write-Host "🔍 Kiểm tra Docker Desktop Kubernetes..." -ForegroundColor Blue
$dockerContext = docker context ls --format "{{.Name}}" | Select-String "desktop-linux"
if ($dockerContext) {
    Write-Host "✅ Docker Desktop context đã sẵn sàng" -ForegroundColor Green
} else {
    Write-Host "⚠️ Docker Desktop context chưa được thiết lập" -ForegroundColor Yellow
}

# Hướng dẫn bật Kubernetes trong Docker Desktop
Write-Host ""
Write-Host "📋 Hướng dẫn bật Kubernetes trong Docker Desktop:" -ForegroundColor Cyan
Write-Host "1. Mở Docker Desktop" -ForegroundColor White
Write-Host "2. Vào Settings (⚙️)" -ForegroundColor White
Write-Host "3. Chọn 'Kubernetes' trong menu bên trái" -ForegroundColor White
Write-Host "4. Tick vào 'Enable Kubernetes'" -ForegroundColor White
Write-Host "5. Chọn 'Apply & Restart'" -ForegroundColor White
Write-Host "6. Chờ Docker Desktop khởi động lại" -ForegroundColor White

Write-Host ""
Write-Host "⏳ Sau khi bật Kubernetes, chạy lại script này để kiểm tra" -ForegroundColor Yellow

# Kiểm tra cluster sau khi bật
Write-Host ""
Write-Host "🔍 Kiểm tra cluster sau khi bật..." -ForegroundColor Blue
try {
    kubectl cluster-info | Out-Null
    Write-Host "✅ Kubernetes cluster đã sẵn sàng!" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "📊 Thông tin cluster:" -ForegroundColor Cyan
    kubectl get nodes
    
    Write-Host ""
    Write-Host "🎉 Bây giờ bạn có thể triển khai ứng dụng:" -ForegroundColor Green
    Write-Host "kubectl apply -f k8s/" -ForegroundColor White
    
} catch {
    Write-Host "❌ Kubernetes cluster chưa sẵn sàng" -ForegroundColor Red
    Write-Host "💡 Vui lòng bật Kubernetes trong Docker Desktop và thử lại" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📚 Tài liệu tham khảo:" -ForegroundColor Cyan
Write-Host "- Docker Desktop Kubernetes: https://docs.docker.com/desktop/kubernetes/" -ForegroundColor White
Write-Host "- kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/" -ForegroundColor White
