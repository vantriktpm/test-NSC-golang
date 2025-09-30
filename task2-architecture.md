# Nhiệm vụ 2: Thiết kế Kiến trúc Nền tảng Bán sản phẩm

## Phân tích Yêu cầu Hệ thống

### Ràng buộc đã cho
- **N = 6000**: Số lượng yêu cầu đồng thời tối đa để xem hoặc mua sản phẩm
- **S = 6**: Số lượng máy chủ tối đa (tối đa 3 có thể được sử dụng làm cơ sở dữ liệu quan hệ)
- **C = 300**: Số lượng kết nối đồng thời tối đa mỗi cơ sở dữ liệu

### Yêu cầu chính
1. **Tính nhất quán dữ liệu**: Không bao giờ cho phép mua hàng thành công vượt quá số lượng có sẵn
2. **Phản hồi thời gian thực**: Cung cấp phản hồi thời gian thực với độ trễ thấp nhất có thể
3. **Tính sẵn sàng cao**: Không có điểm lỗi đơn lẻ (bonus)
4. **Khả năng mở rộng**: Dễ dàng mở rộng khi sản phẩm và người dùng tăng (bonus)

## Thiết kế Kiến trúc Cấp cao

### Sơ đồ Kiến trúc
```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │     (Nginx)     │
                    └─────────┬───────┘
                              │
                    ┌─────────┴───────┐
                    │   CDN/Edge      │
                    │   (CloudFlare)  │
                    └─────────┬───────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │   App   │          │   App   │          │   App   │
   │ Server  │          │ Server  │          │ Server  │
   │   1     │          │   2     │          │   3     │
   └────┬────┘          └────┬────┘          └────┬────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   Redis Cache   │
                    │   (Pub/Sub)     │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │   DB    │          │   DB    │          │   DB    │
   │Primary  │          │Read     │          │Read     │
   │(Master) │          │Replica  │          │Replica  │
   └─────────┘          └─────────┘          └─────────┘
```

## Stack Công nghệ

### Frontend
- **React.js**: Framework UI hiện đại với khả năng thời gian thực
- **WebSocket**: Cập nhật kho và thông báo thời gian thực
- **Redux Toolkit**: Quản lý state cho trạng thái ứng dụng phức tạp
- **Material-UI**: Thư viện component cho thiết kế nhất quán

### Backend
- **Node.js với Express.js**: Runtime JavaScript hiệu suất cao
- **TypeScript**: An toàn kiểu và trải nghiệm phát triển tốt hơn
- **Socket.io**: Giao tiếp hai chiều thời gian thực
- **JWT**: Xác thực và phân quyền

### Tầng Cơ sở dữ liệu
- **PostgreSQL**: Cơ sở dữ liệu quan hệ chính cho tuân thủ ACID
- **Redis**: Tầng cache và pub/sub cho cập nhật thời gian thực
- **Connection Pooling**: pgBouncer cho quản lý kết nối hiệu quả

### Hạ tầng
- **Nginx**: Load balancer và reverse proxy
- **Docker**: Container hóa cho triển khai nhất quán
- **Kubernetes**: Điều phối container (cho khả năng mở rộng)
- **Prometheus + Grafana**: Giám sát và cảnh báo

## Luồng Hoạt động Hệ thống

### 1. Luồng Xem Sản phẩm
```
Yêu cầu Người dùng → Load Balancer → App Server → Redis Cache
                                    ↓
                              Cache Hit? → Trả về Dữ liệu Sản phẩm
                                    ↓
                              Cache Miss → PostgreSQL → Cập nhật Cache → Trả về Dữ liệu
```

### 2. Luồng Mua hàng (Quan trọng cho Tính nhất quán Dữ liệu)
```
Yêu cầu Mua hàng Người dùng → Load Balancer → App Server
                                    ↓
                              Lấy Distributed Lock (Redis)
                                    ↓
                              Kiểm tra Kho trong Database
                                    ↓
                              Kho Có sẵn? → Giảm Kho → Giải phóng Lock
                                    ↓
                              Phát hành Cập nhật Kho (Redis Pub/Sub)
                                    ↓
                              Trả về Phản hồi Thành công
```

### 3. Cập nhật Kho Thời gian thực
```
Thay đổi Kho → Redis Pub/Sub → Tất cả Client Kết nối
                                    ↓
                              Cập nhật UI Thời gian thực
```

## Phân tích Tuân thủ Yêu cầu

### Tính nhất quán Dữ liệu (Yêu cầu Quan trọng)
**Giải pháp**: Distributed locking với optimistic concurrency control

1. **Distributed Lock**: Sử dụng Redis để lấy khóa trên kho sản phẩm
2. **Optimistic Locking**: Cập nhật dựa trên phiên bản trong PostgreSQL
3. **Transaction Isolation**: Sử dụng mức cô lập SERIALIZABLE cho các thao tác quan trọng
4. **Cơ chế Retry**: Exponential backoff cho các giao dịch thất bại

**Triển khai**:
```sql
-- Optimistic locking với trường version
UPDATE products 
SET stock = stock - 1, version = version + 1 
WHERE id = ? AND version = ? AND stock > 0;
```

### Phản hồi Thời gian thực (Yêu cầu Quan trọng)
**Giải pháp**: Kết nối WebSocket với Redis Pub/Sub

1. **Kết nối WebSocket**: Duy trì kết nối liên tục cho cập nhật thời gian thực
2. **Redis Pub/Sub**: Phát sóng thay đổi kho đến tất cả client kết nối
3. **Connection Pooling**: Quản lý kết nối WebSocket hiệu quả
4. **Cơ chế Fallback**: Long polling cho client không thể sử dụng WebSockets

### Tính sẵn sàng cao (Yêu cầu Bonus)
**Giải pháp**: Dự phòng đa tầng

1. **Dự phòng Load Balancer**: Nhiều instance Nginx
2. **Dự phòng Ứng dụng**: 3 app server với health checks
3. **Dự phòng Cơ sở dữ liệu**: Master-slave replication với failover tự động
4. **Dự phòng Cache**: Redis cluster với failover tự động

### Khả năng mở rộng (Yêu cầu Bonus)
**Giải pháp**: Mở rộng ngang với microservices

1. **Ứng dụng Stateless**: Tất cả app server đều stateless
2. **Database Sharding**: Phân vùng sản phẩm theo danh mục hoặc khu vực
3. **Cache Partitioning**: Phân phối dữ liệu Redis trên nhiều node
4. **Tích hợp CDN**: Tài sản tĩnh được phục vụ từ edge locations

## Phân tích Hiệu suất với Số liệu đã cho

### Phân bổ Máy chủ (S = 6)
- **3 App Server**: Xử lý 6000 yêu cầu đồng thời (2000 mỗi server)
- **3 Database Server**: 1 primary + 2 read replica
- **Load Balancer**: Instance riêng biệt (không tính trong S = 6)

### Kết nối Cơ sở dữ liệu (C = 300 mỗi database)
- **Database Primary**: 300 kết nối (ghi + đọc quan trọng)
- **Read Replica 1**: 300 kết nối (thao tác chỉ đọc)
- **Read Replica 2**: 300 kết nối (thao tác chỉ đọc)
- **Tổng**: 900 kết nối cơ sở dữ liệu đồng thời

### Phân bổ Yêu cầu
```
6000 yêu cầu đồng thời → Load Balancer → 3 App Server
                                    ↓
                              ~2000 yêu cầu mỗi server
                                    ↓
                              Database connection pool: 100 mỗi server
                                    ↓
                              Tổng kết nối DB: 300 (trong giới hạn)
```

### Chiến lược Caching
- **Tỷ lệ Cache Hit Redis**: Mục tiêu 90%+ cho dữ liệu sản phẩm
- **Cache TTL**: 5 phút cho thông tin sản phẩm
- **Cập nhật Thời gian thực**: Vô hiệu hóa cache ngay lập tức khi thay đổi kho

## Giám sát và Cảnh báo

### Metrics chính
1. **Thời gian Phản hồi**: < 100ms cho 95th percentile
2. **Throughput**: 6000+ yêu cầu mỗi giây
3. **Tỷ lệ Lỗi**: < 0.1%
4. **Database Connection Pool**: < 80% utilization
5. **Tỷ lệ Cache Hit**: > 90%

### Quy tắc Cảnh báo
- Thời gian phản hồi > 200ms
- Tỷ lệ lỗi > 1%
- Database connection pool > 90%
- Tỷ lệ cache hit < 85%
- Phát hiện không nhất quán kho

## Chiến lược Triển khai

### Môi trường Development
- Single server với Docker Compose
- Instance PostgreSQL và Redis local
- Hot reload cho development

### Môi trường Staging
- 3 app server + 1 database server
- Load balancer cho testing
- Khối lượng dữ liệu giống production

### Môi trường Production
- 3 app server + 3 database server
- Load balancer tính sẵn sàng cao
- CDN cho tài sản tĩnh
- Giám sát và cảnh báo

## Cân nhắc Bảo mật

### Xác thực & Phân quyền
- JWT tokens cho phiên người dùng
- Kiểm soát truy cập dựa trên vai trò (RBAC)
- Giới hạn tốc độ API cho mỗi người dùng

### Bảo vệ Dữ liệu
- HTTPS cho tất cả giao tiếp
- Mã hóa cơ sở dữ liệu khi nghỉ
- Mã hóa dữ liệu PII
- Ngăn chặn SQL injection

### Bảo mật Hạ tầng
- VPC với private subnets
- Security groups cho truy cập mạng
- Cập nhật bảo mật thường xuyên
- Bảo vệ DDoS qua CDN

## Tối ưu Chi phí

### Phân bổ Tài nguyên
- Right-sizing instances dựa trên sử dụng thực tế
- Auto-scaling dựa trên nhu cầu
- Reserved instances cho workload dự đoán được
- Spot instances cho workload không quan trọng

### Tối ưu Cơ sở dữ liệu
- Connection pooling để giảm overhead kết nối
- Read replicas cho thao tác đọc nặng
- Tối ưu query và indexing
- Bảo trì và dọn dẹp thường xuyên

Thiết kế kiến trúc này đảm bảo tất cả yêu cầu được đáp ứng trong khi cung cấp nền tảng vững chắc cho tăng trưởng và khả năng mở rộng trong tương lai.
