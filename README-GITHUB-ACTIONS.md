# Hướng dẫn sử dụng GitHub Actions với API Local

## Tổng quan

GitHub Actions có thể test API local của bạn thông qua ngrok tunnel. Điều này hữu ích để:
- Test API trong môi trường thực tế
- Kiểm tra tích hợp với external services
- Load testing từ GitHub infrastructure
- CI/CD pipeline với local development

## Các bước thực hiện

### 1. Cài đặt ngrok

#### Windows:
```powershell
# Sử dụng Chocolatey
choco install ngrok

# Hoặc tải từ https://ngrok.com/download
```

#### Linux/Mac:
```bash
# Mac với Homebrew
brew install ngrok

# Linux
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
tar -xzf ngrok-v3-stable-linux-amd64.tgz
sudo mv ngrok /usr/local/bin/
```

### 2. Đăng ký và cấu hình ngrok

1. Truy cập https://ngrok.com và đăng ký tài khoản
2. Lấy authtoken từ dashboard
3. Cấu hình ngrok:
```bash
ngrok config add-authtoken YOUR_AUTHTOKEN_HERE
```

### 3. Test API local trước

```powershell
# Windows
.\test-local-api.ps1

# Linux/Mac
./test-local-api.sh
```

### 4. Khởi động ngrok tunnel

```powershell
# Windows - Sử dụng script tự động
.\start-tunnel.ps1

# Hoặc thủ công
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80
ngrok http 8080
```

```bash
# Linux/Mac - Sử dụng script tự động
./start-tunnel.sh

# Hoặc thủ công
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80 &
ngrok http 8080
```

### 5. Lấy ngrok URL

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

**Copy URL `https://abc123.ngrok.io`**

### 6. Chạy GitHub Actions

1. Truy cập repository trên GitHub
2. Vào tab "Actions"
3. Chọn workflow "Local API Testing"
4. Click "Run workflow"
5. Nhập ngrok URL vào field "Local API URL"
6. Chọn test type:
   - `health`: Chỉ test health endpoint
   - `shorten`: Test shorten URL endpoint
   - `analytics`: Test analytics endpoint
   - `load-test`: Chạy load test
   - `all`: Chạy tất cả tests
7. Click "Run workflow"

### 7. Monitor kết quả

- Xem logs trong GitHub Actions tab
- Monitor requests trong ngrok web interface: http://127.0.0.1:4040
- Kiểm tra logs của ứng dụng local

## Workflow Files

### `.github/workflows/local-api-test.yml`

Workflow này cho phép test API local với các tính năng:
- Health check
- URL shortening
- Analytics
- Redirect testing
- Load testing
- Test report generation

### Cách sử dụng:

```yaml
# Trigger workflow manually
on:
  workflow_dispatch:
    inputs:
      local_api_url:
        description: 'Local API URL (e.g., https://abc123.ngrok.io)'
        required: true
        type: string
      test_type:
        description: 'Test type'
        required: true
        default: 'health'
        type: choice
        options:
        - health
        - shorten
        - analytics
        - load-test
        - all
```

## Scripts hỗ trợ

### `start-tunnel.ps1` / `start-tunnel.sh`
- Tự động port forward Kubernetes service
- Khởi động ngrok tunnel
- Cleanup khi thoát

### `test-local-api.ps1` / `test-local-api.sh`
- Test API local trước khi chạy GitHub Actions
- Health check
- URL shortening test
- Analytics test
- Redirect test
- Load test

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
- Kiểm tra firewall/antivirus

### 3. API trả về lỗi
- Kiểm tra ứng dụng local có đang chạy không
- Kiểm tra port forwarding
- Xem logs của ứng dụng
- Kiểm tra database và Redis connection

### 4. Kubernetes service không accessible
```bash
# Kiểm tra pods
kubectl get pods -n url-shortener

# Kiểm tra services
kubectl get services -n url-shortener

# Kiểm tra port forward
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80
```

## Bảo mật

⚠️ **Quan trọng**: ngrok URL là public, ai cũng có thể truy cập được

### Best practices:
- Chỉ sử dụng cho testing/development
- Không expose production data
- Sử dụng ngrok Pro để có custom domain và authentication
- Tắt ngrok khi không sử dụng
- Sử dụng environment variables cho sensitive data

### Ngrok Pro features:
- Custom domains
- Authentication
- IP whitelisting
- Request inspection
- Traffic analysis

## Alternative Solutions

### 1. Cloudflare Tunnel
```bash
# Cài đặt cloudflared
# Windows: choco install cloudflared
# Mac: brew install cloudflared
# Linux: wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64

# Tạo tunnel
cloudflared tunnel --url http://localhost:8080
```

### 2. localtunnel
```bash
# Cài đặt
npm install -g localtunnel

# Sử dụng
lt --port 8080
```

### 3. serveo
```bash
# Sử dụng SSH
ssh -R 80:localhost:8080 serveo.net
```

## Monitoring và Logs

### ngrok Web Interface
- URL: http://127.0.0.1:4040
- Xem tất cả requests
- Inspect request/response
- Traffic analysis

### GitHub Actions Logs
- Real-time logs
- Step-by-step execution
- Error details
- Test results

### Application Logs
```bash
# Kubernetes logs
kubectl logs -f deployment/url-shortener -n url-shortener

# Docker logs
docker logs -f container_name
```

## Performance Testing

### Load Test Configuration
```yaml
- name: Load Test
  run: |
    # Test với 100 requests, 10 concurrent connections
    for i in {1..100}; do
      curl -s "$API_URL/api/v1/health" &
      if (( i % 10 == 0 )); then
        wait
      fi
    done
    wait
```

### Custom Load Test
Bạn có thể tạo custom load test script:

```bash
#!/bin/bash
API_URL="$1"
REQUESTS=${2:-100}
CONCURRENCY=${3:-10}

echo "Testing $REQUESTS requests with $CONCURRENCY concurrent connections..."

for i in $(seq 1 $REQUESTS); do
  curl -s "$API_URL/api/v1/health" > /dev/null &
  if (( i % CONCURRENCY == 0 )); then
    wait
  fi
done
wait

echo "Load test completed"
```

## Kết luận

GitHub Actions với ngrok tunnel là một giải pháp mạnh mẽ để test API local trong môi trường CI/CD. Nó cho phép:

- Test API từ GitHub infrastructure
- Load testing với resources của GitHub
- Integration testing với external services
- CI/CD pipeline hoàn chỉnh

Hãy nhớ luôn bảo mật và chỉ sử dụng cho mục đích testing/development.
