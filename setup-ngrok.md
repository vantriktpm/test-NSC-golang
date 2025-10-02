# Hướng dẫn thiết lập ngrok để GitHub Actions gọi API local

## Bước 1: Cài đặt ngrok

### Windows:
```powershell
# Sử dụng Chocolatey
choco install ngrok

# Hoặc tải trực tiếp từ https://ngrok.com/download
```

### Linux/Mac:
```bash
# Sử dụng Homebrew (Mac)
brew install ngrok

# Hoặc tải trực tiếp
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
tar -xzf ngrok-v3-stable-linux-amd64.tgz
sudo mv ngrok /usr/local/bin/
```

## Bước 2: Đăng ký tài khoản ngrok

1. Truy cập https://ngrok.com
2. Đăng ký tài khoản miễn phí
3. Lấy authtoken từ dashboard

## Bước 3: Cấu hình ngrok

```bash
# Thêm authtoken
ngrok config add-authtoken YOUR_AUTHTOKEN_HERE

# Kiểm tra cấu hình
ngrok config check
```

## Bước 4: Khởi động ngrok tunnel

### Cách 1: Sử dụng ngrok với port forwarding hiện tại
```bash
# Nếu bạn đang port-forward port 8080
ngrok http 8080

# Nếu bạn đang port-forward port 3000 (frontend)
ngrok http 3000
```

### Cách 2: Sử dụng ngrok trực tiếp với Kubernetes
```bash
# Port forward và tạo tunnel
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80 &
ngrok http 8080
```

## Bước 5: Lấy URL public

Sau khi chạy ngrok, bạn sẽ thấy output như:
```
Session Status                online
Account                       your-email@example.com
Version                       3.x.x
Region                        United States (us)
Latency                       -
Web Interface                 http://127.0.0.1:4040
Forwarding                    https://abc123.ngrok.io -> http://localhost:8080
```

**Copy URL `https://abc123.ngrok.io` để sử dụng trong GitHub Actions**

## Bước 6: Chạy GitHub Actions

1. Truy cập repository trên GitHub
2. Vào tab "Actions"
3. Chọn workflow "Local API Testing"
4. Click "Run workflow"
5. Nhập ngrok URL vào field "Local API URL"
6. Chọn test type và click "Run workflow"

## Bước 7: Monitor kết quả

- Xem logs trong GitHub Actions
- Monitor requests trong ngrok web interface: http://127.0.0.1:4040
- Kiểm tra logs của ứng dụng local

## Troubleshooting

### 1. ngrok không kết nối được
```bash
# Kiểm tra authtoken
ngrok config check

# Kiểm tra kết nối internet
ping ngrok.com
```

### 2. GitHub Actions không gọi được API
- Đảm bảo ngrok tunnel vẫn đang chạy
- Kiểm tra URL ngrok có đúng không
- Xem logs trong ngrok web interface

### 3. API trả về lỗi
- Kiểm tra ứng dụng local có đang chạy không
- Kiểm tra port forwarding
- Xem logs của ứng dụng

## Lưu ý bảo mật

⚠️ **Quan trọng**: ngrok URL là public, ai cũng có thể truy cập được

- Chỉ sử dụng cho testing/development
- Không expose production data
- Sử dụng ngrok Pro để có custom domain và authentication
- Tắt ngrok khi không sử dụng

## Alternative: Sử dụng Cloudflare Tunnel

Nếu không muốn sử dụng ngrok, có thể dùng Cloudflare Tunnel:

```bash
# Cài đặt cloudflared
# Windows: choco install cloudflared
# Mac: brew install cloudflared
# Linux: wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64

# Tạo tunnel
cloudflared tunnel --url http://localhost:8080
```

## Script tự động hóa

Tạo file `start-tunnel.ps1` (Windows):

```powershell
# Start tunnel script
Write-Host "🚀 Starting ngrok tunnel..." -ForegroundColor Green

# Check if ngrok is installed
if (!(Get-Command ngrok -ErrorAction SilentlyContinue)) {
    Write-Host "❌ ngrok not found. Please install ngrok first." -ForegroundColor Red
    exit 1
}

# Start port forward in background
Write-Host "📡 Starting port forward..." -ForegroundColor Blue
Start-Job -ScriptBlock {
    kubectl port-forward -n url-shortener service/url-shortener-service 8080:80
}

# Wait a bit for port forward to start
Start-Sleep -Seconds 5

# Start ngrok
Write-Host "🌐 Starting ngrok tunnel..." -ForegroundColor Blue
ngrok http 8080
```

Tạo file `start-tunnel.sh` (Linux/Mac):

```bash
#!/bin/bash
echo "🚀 Starting ngrok tunnel..."

# Check if ngrok is installed
if ! command -v ngrok &> /dev/null; then
    echo "❌ ngrok not found. Please install ngrok first."
    exit 1
fi

# Start port forward in background
echo "📡 Starting port forward..."
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80 &
PORT_FORWARD_PID=$!

# Wait a bit for port forward to start
sleep 5

# Start ngrok
echo "🌐 Starting ngrok tunnel..."
ngrok http 8080

# Cleanup
kill $PORT_FORWARD_PID
```
