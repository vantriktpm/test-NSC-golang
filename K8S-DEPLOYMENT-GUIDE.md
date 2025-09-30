# Hướng dẫn triển khai URL Shortener lên Kubernetes

## Yêu cầu hệ thống

1. **Kubernetes cluster** đang chạy (minikube, Docker Desktop, hoặc cloud cluster)
2. **kubectl** đã được cài đặt và cấu hình
3. **Docker** đã được cài đặt

## Các bước triển khai

### 1. Kiểm tra cluster

```bash
# Kiểm tra kết nối cluster
kubectl cluster-info

# Kiểm tra nodes
kubectl get nodes
```

### 2. Build Docker image

```bash
# Build image
docker build -t url-shortener:latest .

# Nếu sử dụng minikube, load image vào minikube
minikube image load url-shortener:latest
```

### 3. Triển khai ứng dụng

#### Cách 1: Sử dụng script tự động

```bash
# Chạy script triển khai
./deploy-k8s.sh
```

#### Cách 2: Triển khai thủ công

```bash
# 1. Tạo namespace
kubectl apply -f k8s/namespace.yaml

# 2. Triển khai PostgreSQL
kubectl apply -f k8s/postgres-deployment.yaml

# 3. Triển khai Redis
kubectl apply -f k8s/redis-deployment.yaml

# 4. Chờ database sẵn sàng
kubectl wait --for=condition=ready pod -l app=postgres -n url-shortener --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n url-shortener --timeout=300s

# 5. Tạo secrets và configmap
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/configmap.yaml

# 6. Triển khai ứng dụng chính
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 7. Chờ ứng dụng sẵn sàng
kubectl wait --for=condition=ready pod -l app=url-shortener -n url-shortener --timeout=300s

# 8. Tạo Ingress (tùy chọn)
kubectl apply -f k8s/ingress.yaml
```

### 4. Kiểm tra trạng thái

```bash
# Xem pods
kubectl get pods -n url-shortener

# Xem services
kubectl get services -n url-shortener

# Xem ingress
kubectl get ingress -n url-shortener

# Xem logs
kubectl logs -f deployment/url-shortener -n url-shortener
```

### 5. Truy cập ứng dụng

#### Cách 1: Port Forward

```bash
# Port forward service
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80

# Truy cập: http://localhost:8080
```

#### Cách 2: Sử dụng Ingress

Nếu đã cấu hình Ingress controller:

```bash
# Lấy IP của ingress
kubectl get ingress -n url-shortener

# Truy cập qua IP hoặc domain đã cấu hình
```

### 6. Test ứng dụng

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Shorten URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Redirect test (sử dụng short code từ response trên)
curl -I http://localhost:8080/{short_code}

# Analytics
curl http://localhost:8080/api/v1/analytics/{short_code}
```

## Cấu trúc file Kubernetes

```
k8s/
├── namespace.yaml          # Namespace cho ứng dụng
├── postgres-deployment.yaml # PostgreSQL database
├── redis-deployment.yaml   # Redis cache
├── deployment.yaml         # Ứng dụng chính
├── service.yaml           # Service cho ứng dụng
├── configmap.yaml         # Cấu hình ứng dụng
├── secret.yaml           # Secrets (database, redis)
└── ingress.yaml          # Ingress controller
```

## Troubleshooting

### 1. Pod không start

```bash
# Xem logs
kubectl logs -f pod/{pod-name} -n url-shortener

# Xem events
kubectl get events -n url-shortener --sort-by='.lastTimestamp'
```

### 2. Database connection issues

```bash
# Kiểm tra PostgreSQL
kubectl exec -it deployment/postgres -n url-shortener -- psql -U user -d urlshortener

# Kiểm tra Redis
kubectl exec -it deployment/redis -n url-shortener -- redis-cli ping
```

### 3. Service không accessible

```bash
# Kiểm tra endpoints
kubectl get endpoints -n url-shortener

# Test service từ trong cluster
kubectl run test-pod --image=busybox -it --rm --restart=Never -- nslookup url-shortener-service.url-shortener.svc.cluster.local
```

## Scaling

```bash
# Scale ứng dụng
kubectl scale deployment url-shortener --replicas=5 -n url-shortener

# Auto scaling (cần cài đặt HPA)
kubectl autoscale deployment url-shortener --cpu-percent=70 --min=3 --max=10 -n url-shortener
```

## Cleanup

```bash
# Xóa tất cả resources
kubectl delete namespace url-shortener

# Hoặc xóa từng resource
kubectl delete -f k8s/
```

## Monitoring

```bash
# Xem resource usage
kubectl top pods -n url-shortener
kubectl top nodes

# Xem metrics
kubectl get --raw /apis/metrics.k8s.io/v1beta1/pods
```

## Production Considerations

1. **Persistent Volumes**: Đảm bảo data được lưu trữ bền vững
2. **Resource Limits**: Thiết lập limits phù hợp cho production
3. **Health Checks**: Cấu hình liveness và readiness probes
4. **Security**: Sử dụng NetworkPolicies, PodSecurityPolicies
5. **Monitoring**: Cài đặt Prometheus, Grafana
6. **Logging**: Cấu hình centralized logging
7. **Backup**: Thiết lập backup cho database
8. **SSL/TLS**: Cấu hình HTTPS cho production
