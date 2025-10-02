# Hướng dẫn truy cập Frontend trên Kubernetes

## Các cách truy cập Frontend sau khi deploy lên K8s

### 1. Port Forward (Development/Testing)

Port forwarding là cách đơn giản nhất để truy cập ứng dụng từ local:

```bash
# Port forward frontend service
kubectl port-forward -n url-shortener service/url-shortener-frontend-service 3000:80

# Port forward backend service (nếu cần)
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80
```

**Truy cập:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

**Lưu ý:** 
- Port forward chỉ dành cho development/testing
- Khi đóng terminal, connection sẽ bị ngắt
- Chỉ một người có thể truy cập tại một thời điểm

### 2. NodePort Service

Expose service qua NodePort để truy cập từ bất kỳ node nào trong cluster:

```bash
# Chuyển service sang NodePort
kubectl patch service url-shortener-frontend-service -n url-shortener -p '{"spec":{"type":"NodePort"}}'

# Xem port được assign
kubectl get service url-shortener-frontend-service -n url-shortener
```

**Truy cập:**
```bash
# Lấy IP của node
kubectl get nodes -o wide

# Truy cập: http://<NODE-IP>:<NODE-PORT>
# Với Docker Desktop: http://localhost:<NODE-PORT>
```

### 3. LoadBalancer (Cloud/Production)

Sử dụng LoadBalancer khi deploy trên cloud (AWS, GCP, Azure):

```bash
# Chuyển service sang LoadBalancer
kubectl patch service url-shortener-frontend-service -n url-shortener -p '{"spec":{"type":"LoadBalancer"}}'

# Xem external IP
kubectl get service url-shortener-frontend-service -n url-shortener
```

**Truy cập:**
- Cloud sẽ tự động provision một External IP/DNS
- Truy cập: http://<EXTERNAL-IP>

**Lưu ý:** LoadBalancer không hoạt động với Docker Desktop hoặc Minikube local

### 4. Ingress Controller (Production - Khuyến nghị)

Sử dụng Ingress để có nhiều tính năng như SSL/TLS, routing, load balancing:

#### Bước 1: Cài đặt Ingress Controller

**Với Docker Desktop:**
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
```

**Với Minikube:**
```bash
minikube addons enable ingress
```

#### Bước 2: Kiểm tra Ingress Controller

```bash
kubectl get pods -n ingress-nginx
kubectl get services -n ingress-nginx
```

#### Bước 3: Apply Ingress Resource

Ingress đã được cấu hình trong `k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: url-shortener-ingress
  namespace: url-shortener
spec:
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: url-shortener-service
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: url-shortener-frontend-service
            port:
              number: 80
```

#### Bước 4: Cấu hình hosts file (Local testing)

**Windows:** `C:\Windows\System32\drivers\etc\hosts`
**Linux/Mac:** `/etc/hosts`

Thêm dòng:
```
127.0.0.1 your-domain.com
```

Hoặc sử dụng domain thật và DNS nếu deploy production.

#### Bước 5: Truy cập

```bash
# Xem ingress
kubectl get ingress -n url-shortener

# Truy cập
# http://your-domain.com (frontend)
# http://your-domain.com/api (backend)
```

### 5. Sử dụng Minikube Tunnel (chỉ Minikube)

Nếu sử dụng Minikube và muốn test LoadBalancer:

```bash
# Mở terminal riêng và chạy
minikube tunnel

# Trong terminal khác, patch service
kubectl patch service url-shortener-frontend-service -n url-shortener -p '{"spec":{"type":"LoadBalancer"}}'

# Truy cập: http://localhost
```

## Cấu hình đã triển khai

### Services hiện tại:

```bash
kubectl get services -n url-shortener
```

Output:
```
NAME                             TYPE        CLUSTER-IP      PORT(S)
postgres-service                 ClusterIP   10.110.39.125   5432/TCP
redis-service                    ClusterIP   10.101.25.187   6379/TCP
url-shortener-service            ClusterIP   10.102.65.221   80/TCP
url-shortener-frontend-service   ClusterIP   10.99.240.255   80/TCP
```

### Pods hiện tại:

```bash
kubectl get pods -n url-shortener
```

Output:
```
NAME                                     READY   STATUS
postgres-7dc7f758db-ggd2l                1/1     Running
redis-7f878f58cf-j8cgp                   1/1     Running
url-shortener-666868cb57-dsj6v           1/1     Running
url-shortener-666868cb57-hdgfg           1/1     Running
url-shortener-666868cb57-zc48b           1/1     Running
url-shortener-frontend-d858b7d95-dc669   1/1     Running
url-shortener-frontend-d858b7d95-znt9w   1/1     Running
```

## Kiểm tra trạng thái

### Health check:

```bash
# Frontend health
curl http://localhost:3000/health

# Backend health
curl http://localhost:8080/api/v1/health
```

### Xem logs:

```bash
# Frontend logs
kubectl logs -f -l app=url-shortener-frontend -n url-shortener

# Backend logs
kubectl logs -f -l app=url-shortener -n url-shortener
```

### Truy cập vào container:

```bash
# Vào frontend container
kubectl exec -it deployment/url-shortener-frontend -n url-shortener -- /bin/sh

# Vào backend container
kubectl exec -it deployment/url-shortener -n url-shortener -- /bin/sh
```

## Troubleshooting

### 1. Frontend không load được

```bash
# Kiểm tra pods
kubectl get pods -n url-shortener -l app=url-shortener-frontend

# Xem logs
kubectl logs -l app=url-shortener-frontend -n url-shortener

# Xem chi tiết pod
kubectl describe pod -l app=url-shortener-frontend -n url-shortener
```

### 2. API calls từ frontend bị lỗi

Kiểm tra nginx config trong frontend container:
```bash
kubectl exec -it deployment/url-shortener-frontend -n url-shortener -- cat /etc/nginx/nginx.conf
```

Đảm bảo proxy_pass trỏ đúng:
```nginx
location /api/ {
    proxy_pass http://url-shortener-service:80;
}
```

### 3. Cannot access từ browser

- Kiểm tra port forward vẫn running
- Kiểm tra firewall/antivirus
- Thử browser khác hoặc incognito mode
- Clear browser cache

### 4. DNS resolution issues trong cluster

Test DNS từ một pod:
```bash
kubectl run test-dns --image=busybox -it --rm --restart=Never -- nslookup url-shortener-service.url-shortener.svc.cluster.local
```

## Best Practices

### Development:
- Sử dụng Port Forward cho testing nhanh
- Log vào pods để debug
- Sử dụng `kubectl port-forward` với nhiều terminal

### Staging:
- Sử dụng NodePort hoặc Ingress
- Cấu hình DNS nội bộ
- Monitor logs và metrics

### Production:
- **Khuyến nghị**: Sử dụng Ingress Controller với SSL/TLS
- Cấu hình domain và DNS chính thức
- Bật monitoring (Prometheus, Grafana)
- Setup backup và disaster recovery
- Cấu hình auto-scaling
- Implement rate limiting và security measures

## Tổng kết các lệnh cần nhớ

```bash
# Deploy frontend
kubectl apply -f k8s/frontend-deployment.yaml

# Port forward
kubectl port-forward -n url-shortener service/url-shortener-frontend-service 3000:80

# Xem trạng thái
kubectl get all -n url-shortener

# Xem logs
kubectl logs -f deployment/url-shortener-frontend -n url-shortener

# Restart deployment
kubectl rollout restart deployment/url-shortener-frontend -n url-shortener

# Delete và redeploy
kubectl delete deployment url-shortener-frontend -n url-shortener
kubectl apply -f k8s/frontend-deployment.yaml
```

