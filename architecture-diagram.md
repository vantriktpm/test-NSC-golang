# Sơ đồ Kiến trúc Hệ thống

## Kiến trúc Cấp cao

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

## Sơ đồ Luồng Yêu cầu

```
Yêu cầu Người dùng
     │
     ▼
┌─────────────┐
│ Load        │
│ Balancer    │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ App Server  │
│ (Round      │
│ Robin)      │
└─────┬───────┘
      │
      ▼
┌─────────────┐    Cache Hit? ──► Trả về Dữ liệu
│ Redis       │
│ Cache       │
└─────┬───────┘
      │ Cache Miss
      ▼
┌─────────────┐
│ PostgreSQL  │
│ Database    │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Cập nhật    │
│ Cache       │
└─────────────┘
```

## Sơ đồ Luồng Mua hàng

```
Yêu cầu Mua hàng
     │
     ▼
┌─────────────┐
│ App Server  │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Lấy         │
│ Distributed │
│ Lock        │
│ (Redis)     │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Kiểm tra    │
│ Kho trong   │
│ Database    │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Kho         │
│ Có sẵn?     │
└─────┬───────┘
      │ Có
      ▼
┌─────────────┐
│ Giảm        │
│ Kho         │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Giải phóng  │
│ Lock        │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Phát hành   │
│ Cập nhật    │
│ Kho         │
│ (Pub/Sub)   │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Trả về      │
│ Thành công  │
└─────────────┘
```

## Luồng Cập nhật Thời gian thực

```
Sự kiện Thay đổi Kho
     │
     ▼
┌─────────────┐
│ PostgreSQL  │
│ Trigger     │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Redis       │
│ Pub/Sub     │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ WebSocket   │
│ Connections │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Client      │
│ Cập nhật UI │
└─────────────┘
```

## Phân bổ Máy chủ (S = 6)

```
Máy chủ 1: App Server (Node.js + Express)
Máy chủ 2: App Server (Node.js + Express)  
Máy chủ 3: App Server (Node.js + Express)
Máy chủ 4: Database Primary (PostgreSQL)
Máy chủ 5: Database Read Replica (PostgreSQL)
Máy chủ 6: Database Read Replica (PostgreSQL)

Bổ sung (không tính trong S=6):
- Load Balancer (Nginx)
- Redis Cache (có thể chạy trên app servers)
- CDN (CloudFlare - external)
```

## Phân bổ Kết nối Cơ sở dữ liệu (C = 300 mỗi DB)

```
App Server 1 ──► DB Primary (100 kết nối)
App Server 2 ──► DB Primary (100 kết nối)  
App Server 3 ──► DB Primary (100 kết nối)
                Tổng: 300 kết nối ✅

App Server 1 ──► DB Replica 1 (100 kết nối)
App Server 2 ──► DB Replica 1 (100 kết nối)
App Server 3 ──► DB Replica 1 (100 kết nối)
                Tổng: 300 kết nối ✅

App Server 1 ──► DB Replica 2 (100 kết nối)
App Server 2 ──► DB Replica 2 (100 kết nối)
App Server 3 ──► DB Replica 2 (100 kết nối)
                Tổng: 300 kết nối ✅
```

## Phân bổ Yêu cầu Đồng thời (N = 6000)

```
6000 yêu cầu đồng thời
     │
     ▼
┌─────────────┐
│ Load        │
│ Balancer    │
│ (Nginx)     │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ App Server 1│ ──► 2000 yêu cầu
│ App Server 2│ ──► 2000 yêu cầu  
│ App Server 3│ ──► 2000 yêu cầu
└─────────────┘
                Tổng: 6000 yêu cầu ✅
```
