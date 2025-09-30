# Hướng dẫn Test 6000 Requests cho API URL Shortener

## Tổng quan

Dự án này cung cấp nhiều cách để test 6000 requests đồng thời cho API URL Shortener, bao gồm:

- **Bash script** với Apache Bench (ab) hoặc wrk
- **Go script** với goroutines
- **Python script** với asyncio và aiohttp
- **Node.js script** với async/await

## Yêu cầu

### 1. Chuẩn bị môi trường

```bash
# Khởi động API server
docker-compose up -d

# Hoặc chạy trực tiếp
go run main.go
```

### 2. Cài đặt công cụ test

#### Apache Bench (ab)
```bash
# Ubuntu/Debian
sudo apt-get install apache2-utils

# macOS
brew install httpd

# CentOS/RHEL
sudo yum install httpd-tools
```

#### wrk
```bash
# Ubuntu/Debian
sudo apt-get install wrk

# macOS
brew install wrk

# Từ source
git clone https://github.com/wg/wrk.git
cd wrk
make
```

## Cách sử dụng

### 1. Bash Script (Apache Bench/wrk)

```bash
# Cấp quyền thực thi
chmod +x test-6000-requests.sh

# Chạy test
./test-6000-requests.sh
```

**Tùy chỉnh:**
```bash
# Thay đổi số lượng requests
TOTAL_REQUESTS=10000 ./test-6000-requests.sh

# Thay đổi số concurrent users
CONCURRENT_USERS=200 ./test-6000-requests.sh

# Thay đổi base URL
BASE_URL="http://localhost:3000" ./test-6000-requests.sh
```

### 2. Go Script

```bash
# Chạy trực tiếp
go run test-6000-requests-go.go

# Hoặc build và chạy
go build -o load-test test-6000-requests-go.go
./load-test
```

**Tùy chỉnh:**
```go
const (
    baseURL        = "http://localhost:8080"
    totalRequests  = 6000
    concurrentUsers = 100
)
```

### 3. Python Script

```bash
# Cài đặt dependencies
pip install -r requirements.txt

# Chạy test
python test-6000-requests-python.py
```

**Tùy chỉnh:**
```python
BASE_URL = "http://localhost:8080"
TOTAL_REQUESTS = 6000
CONCURRENT_USERS = 100
```

### 4. Node.js Script

```bash
# Cài đặt dependencies (nếu cần)
npm install

# Chạy test
node test-6000-requests-node.js

# Hoặc sử dụng npm script
npm test
```

**Tùy chỉnh:**
```javascript
const BASE_URL = 'http://localhost:8080';
const TOTAL_REQUESTS = 6000;
const CONCURRENT_USERS = 100;
```

## Các endpoint được test

1. **Health Check**: `GET /api/v1/health`
2. **Shorten URL**: `POST /api/v1/shorten`
3. **Redirect URL**: `GET /{shortCode}`
4. **Analytics**: `GET /api/v1/analytics/{shortCode}`

## Kết quả mong đợi

### Metrics được đo lường:
- **Total requests**: Tổng số requests
- **Success rate**: Tỷ lệ thành công
- **Response time**: Thời gian phản hồi (min, max, avg, median, p95, p99)
- **Requests per second**: Số requests mỗi giây
- **Total time**: Tổng thời gian test

### Kết quả mẫu:
```
Endpoint: /api/v1/health
Total requests: 6000
Success: 6000 (100.00%)
Errors: 0 (0.00%)
Total time: 15.23s
Requests/sec: 393.96
Min response time: 0.001s
Max response time: 0.045s
Avg response time: 0.012s
Median response time: 0.010s
95th percentile: 0.025s
99th percentile: 0.035s
```

## Tối ưu hóa hiệu suất

### 1. Tăng số concurrent users
```bash
# Test với 200 concurrent users
CONCURRENT_USERS=200 ./test-6000-requests.sh
```

### 2. Test với số lượng requests lớn hơn
```bash
# Test với 10000 requests
TOTAL_REQUESTS=10000 ./test-6000-requests.sh
```

### 3. Test trên nhiều endpoint cùng lúc
```bash
# Chạy test song song cho nhiều endpoint
./test-6000-requests.sh &
./test-6000-requests.sh &
wait
```

## Troubleshooting

### 1. Lỗi "Connection refused"
```bash
# Kiểm tra API server có chạy không
curl http://localhost:8080/api/v1/health

# Khởi động API server
docker-compose up -d
```

### 2. Lỗi "Too many open files"
```bash
# Tăng giới hạn file descriptors
ulimit -n 65536

# Hoặc thêm vào ~/.bashrc
echo "ulimit -n 65536" >> ~/.bashrc
```

### 3. Lỗi timeout
```bash
# Tăng timeout trong script
TIMEOUT=60 ./test-6000-requests.sh
```

### 4. Lỗi memory
```bash
# Giảm số concurrent users
CONCURRENT_USERS=50 ./test-6000-requests.sh
```

## So sánh các công cụ

| Công cụ | Ưu điểm | Nhược điểm | Phù hợp cho |
|---------|---------|------------|-------------|
| Apache Bench | Đơn giản, nhanh | Ít tính năng | Test cơ bản |
| wrk | Hiệu suất cao | Cần script Lua | Test chuyên sâu |
| Go | Kiểm soát tốt | Cần build | Custom logic |
| Python | Dễ sử dụng | Chậm hơn | Prototyping |
| Node.js | Async tốt | Memory usage | Web apps |

## Kết luận

Chọn công cụ phù hợp với nhu cầu:
- **Apache Bench**: Test nhanh và đơn giản
- **wrk**: Test hiệu suất cao
- **Go**: Custom logic phức tạp
- **Python**: Prototyping và analysis
- **Node.js**: Test web applications
