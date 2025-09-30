# Bài Tập Fullstack Engineer

## Nhiệm vụ 1: Dịch vụ Rút gọn URL với Phân tích

### Yêu cầu và Thông số kỹ thuật

#### Tính năng cốt lõi
- **Rút gọn URL**: Chuyển đổi URL dài thành liên kết ngắn, có thể chia sẻ
- **Chuyển hướng URL**: Chuyển hướng URL ngắn về URL gốc
- **Phân tích**: Theo dõi thống kê lượt click và các chỉ số sử dụng cơ bản
- **API**: RESTful API cho tất cả các thao tác

#### Yêu cầu kỹ thuật
- **Backend**: Golang (sử dụng framework Gin)
- **Frontend**: Vue 3 với TypeScript và Vite
- **Cơ sở dữ liệu**: PostgreSQL để lưu trữ dữ liệu
- **Bộ nhớ đệm**: Redis để tra cứu URL hiệu suất cao
- **Container hóa**: Docker với docker-compose
- **CI/CD**: GitHub Actions cho linting và testing
- **Tùy chọn**: Cấu hình triển khai Kubernetes

#### Các endpoint API
```
POST /api/v1/shorten     - Tạo URL ngắn
GET  /{shortCode}        - Chuyển hướng về URL gốc
GET  /api/v1/analytics/{shortCode} - Lấy phân tích cho URL ngắn
GET  /api/v1/health      - Kiểm tra sức khỏe
```

#### Schema cơ sở dữ liệu
```sql
-- Bảng URLs
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Bảng Analytics
CREATE TABLE analytics (
    id SERIAL PRIMARY KEY,
    url_id INTEGER REFERENCES urls(id),
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Kiến trúc
- **Frontend**: Vue 3 SPA với responsive design
- **Web Server**: Gin HTTP server
- **Cơ sở dữ liệu**: PostgreSQL để lưu trữ
- **Bộ nhớ đệm**: Redis để tra cứu URL nhanh
- **Load Balancer**: Nginx (cho production)
- **Giám sát**: Kiểm tra sức khỏe cơ bản và metrics

## Nhiệm vụ 2: Kiến trúc Nền tảng Bán sản phẩm

### Yêu cầu hệ thống
- **Người dùng đồng thời**: 6000 yêu cầu đồng thời
- **Máy chủ**: Tối đa 6 máy chủ (3 cho cơ sở dữ liệu)
- **Kết nối cơ sở dữ liệu**: 300 kết nối đồng thời mỗi cơ sở dữ liệu
- **Tính nhất quán dữ liệu**: Ngăn chặn bán quá số lượng (nhất quán kho)
- **Phản hồi thời gian thực**: Phản hồi độ trễ thấp
- **Tính sẵn sàng cao**: Không có điểm lỗi đơn lẻ
- **Khả năng mở rộng**: Dễ dàng mở rộng khi tăng trưởng

### Stack công nghệ
- **Frontend**: React.js với cập nhật thời gian thực
- **Backend**: Node.js với Express.js
- **Cơ sở dữ liệu**: PostgreSQL (chính) + Redis (cache)
- **Hàng đợi tin nhắn**: Redis Pub/Sub cho cập nhật thời gian thực
- **Load Balancer**: Nginx
- **CDN**: CloudFlare cho tài sản tĩnh
- **Giám sát**: Prometheus + Grafana

### Thiết kế kiến trúc
Xem `architecture-diagram.md` để biết sơ đồ chi tiết và `task2-architecture.md` để biết tài liệu kiến trúc đầy đủ.

### Luồng hệ thống
1. **Xem sản phẩm**: Thao tác đọc nhanh với Redis caching
2. **Luồng mua hàng**: Distributed locking để ngăn chặn bán quá số lượng
3. **Cập nhật thời gian thực**: Kết nối WebSocket cho cập nhật kho trực tiếp
4. **Tính nhất quán dữ liệu**: Optimistic locking với cơ chế retry

### Tính năng chính
- **Distributed Locking**: Khóa dựa trên Redis cho quản lý kho
- **Cập nhật thời gian thực**: WebSocket + Redis Pub/Sub cho thay đổi kho trực tiếp
- **Tính sẵn sàng cao**: Dự phòng đa tầng với failover tự động
- **Khả năng mở rộng**: Mở rộng ngang với máy chủ ứng dụng stateless
- **Hiệu suất**: Tỷ lệ cache hit 90%+ với thời gian phản hồi dưới 100ms

## Bắt đầu

### Yêu cầu tiên quyết
- Docker và Docker Compose
- Go 1.21+ (cho URL shortener backend)
- Node.js 18+ (cho frontend và nền tảng sản phẩm)
- PostgreSQL 14+
- Redis 7+

### Cài đặt và Chạy

#### Dịch vụ Rút gọn URL

##### Full Stack với Docker Compose
```bash
# Clone repository
git clone <repository-url>
cd url-shortener

# Khởi động tất cả services (Backend + Frontend + Database + Redis)
docker-compose up --build -d

# Kiểm tra logs
docker-compose logs -f

# Dừng services
docker-compose down
```

**Services sẽ chạy trên:**
- **Backend API**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

##### Backend (Golang) - Development
```bash
# Chạy tests
go test ./...

# Chạy linting
golangci-lint run

# Chạy load test
go run test-6000-requests-go.go
```

##### Frontend (Vue 3) - Development
```bash
# Di chuyển vào thư mục frontend
cd frontend

# Cài đặt dependencies
npm install

# Khởi động development server
npm run dev

# Build cho production
npm run build

# Chạy linting
npm run lint
```

##### Makefile Commands
```bash
# Khởi động full stack
make docker-compose-up

# Dừng services
make docker-compose-down

# Xem logs
make docker-compose-logs

# Restart services
make docker-compose-restart

# Build frontend
make frontend-build

# Chạy frontend development
make frontend-dev

# Chạy load test
make load-test

# Xem tất cả commands
make help
```

##### Tính năng Frontend
- **URL Shortener**: Giao diện rút gọn URL với copy to clipboard
- **Analytics Dashboard**: Biểu đồ thống kê và phân tích chi tiết
- **Health Check**: Theo dõi sức khỏe hệ thống real-time
- **Load Testing**: Công cụ kiểm thử tải với cấu hình linh hoạt
- **Bulk Operations**: Thao tác hàng loạt với export CSV
- **Responsive Design**: Tối ưu cho mobile và desktop

#### Nền tảng Sản phẩm
```bash
# Cài đặt dependencies
npm install

# Khởi động development server
npm run dev

# Chạy tests
npm test

# Build cho production
npm run build
```

## Triển khai

### Triển khai Docker

#### Backend và Database
```bash
# Build và chạy với Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

#### Frontend
```bash
# Build Docker image cho frontend
cd frontend
docker build -t url-shortener-frontend .

# Chạy container
docker run -p 80:80 url-shortener-frontend
```

### Triển khai Kubernetes

#### Thiết lập Kubernetes Local
```bash
# Thiết lập cluster local (Windows)
.\setup-k8s-local.ps1

# Triển khai ứng dụng
.\deploy-k8s-local.ps1
```

#### Triển khai thủ công
```bash
# Áp dụng cấu hình Kubernetes
kubectl apply -f k8s/

# Kiểm tra trạng thái
kubectl get pods -n url-shortener
kubectl get services -n url-shortener
```

**Lưu ý**: Xem `K8S-LOCAL-SETUP.md` để biết hướng dẫn chi tiết thiết lập Kubernetes local.

## Giám sát và Phân tích

### Kiểm tra sức khỏe
- URL Shortener Backend: `GET /api/v1/health`
- URL Shortener Frontend: `http://localhost:3000` (development) hoặc `http://localhost:80` (production)
- Nền tảng Sản phẩm: `GET /api/health`

### Metrics
- Số lượng yêu cầu và thời gian phản hồi
- Trạng thái connection pool cơ sở dữ liệu
- Tỷ lệ cache hit/miss
- Tỷ lệ lỗi và loại lỗi
- Frontend performance metrics (load time, bundle size)
- User interaction analytics (click tracking, usage patterns)

## Đóng góp

1. Fork repository
2. Tạo feature branch
3. Thực hiện thay đổi
4. Thêm tests
5. Gửi pull request

## Giấy phép

MIT License
