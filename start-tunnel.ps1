# Script khởi động ngrok tunnel cho GitHub Actions testing

param(
    [string]$Port = "8080",
    [string]$Service = "url-shortener-service",
    [string]$Namespace = "url-shortener"
)

Write-Host "🚀 Starting ngrok tunnel for GitHub Actions..." -ForegroundColor Green

# Check if ngrok is installed
if (!(Get-Command ngrok -ErrorAction SilentlyContinue)) {
    Write-Host "❌ ngrok not found. Please install ngrok first." -ForegroundColor Red
    Write-Host "💡 Install with: choco install ngrok" -ForegroundColor Yellow
    Write-Host "💡 Or download from: https://ngrok.com/download" -ForegroundColor Yellow
    exit 1
}

# Check if kubectl is available
if (!(Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Host "❌ kubectl not found. Please install kubectl first." -ForegroundColor Red
    exit 1
}

# Check if Kubernetes cluster is running
try {
    kubectl cluster-info | Out-Null
    Write-Host "✅ Kubernetes cluster is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Kubernetes cluster not accessible" -ForegroundColor Red
    Write-Host "💡 Make sure Docker Desktop Kubernetes is enabled" -ForegroundColor Yellow
    exit 1
}

# Check if service exists
try {
    kubectl get service $Service -n $Namespace | Out-Null
    Write-Host "✅ Service $Service found in namespace $Namespace" -ForegroundColor Green
} catch {
    Write-Host "❌ Service $Service not found in namespace $Namespace" -ForegroundColor Red
    Write-Host "💡 Available services:" -ForegroundColor Yellow
    kubectl get services -n $Namespace
    exit 1
}

Write-Host ""
Write-Host "📋 Configuration:" -ForegroundColor Cyan
Write-Host "  Port: $Port" -ForegroundColor White
Write-Host "  Service: $Service" -ForegroundColor White
Write-Host "  Namespace: $Namespace" -ForegroundColor White
Write-Host ""

# Start port forward in background
Write-Host "📡 Starting port forward..." -ForegroundColor Blue
$portForwardJob = Start-Job -ScriptBlock {
    param($Service, $Namespace, $Port)
    kubectl port-forward -n $Namespace service/$Service $Port`:80
} -ArgumentList $Service, $Namespace, $Port

# Wait for port forward to start
Write-Host "⏳ Waiting for port forward to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check if port forward is working
try {
    $response = Invoke-WebRequest -Uri "http://localhost:$Port/api/v1/health" -Method GET -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Port forward is working" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Port forward might not be ready yet, continuing..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🌐 Starting ngrok tunnel..." -ForegroundColor Blue
Write-Host "💡 Copy the HTTPS URL from ngrok output for GitHub Actions" -ForegroundColor Yellow
Write-Host "💡 Press Ctrl+C to stop the tunnel" -ForegroundColor Yellow
Write-Host ""

# Start ngrok
try {
    ngrok http $Port
} finally {
    # Cleanup: Stop port forward job
    Write-Host ""
    Write-Host "🧹 Cleaning up..." -ForegroundColor Blue
    Stop-Job $portForwardJob -ErrorAction SilentlyContinue
    Remove-Job $portForwardJob -ErrorAction SilentlyContinue
    Write-Host "✅ Cleanup completed" -ForegroundColor Green
}
