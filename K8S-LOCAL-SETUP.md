# Hướng dẫn thiết lập Kubernetes Local

## Vấn đề hiện tại

Lỗi `kubectl apply -f k8s/` không thành công vì:
- Kubernetes cluster chưa được thiết lập
- Docker Desktop chưa bật Kubernetes
- kubectl chưa được cấu hình đúng

## Giải pháp

### Bước 1: Thiết lập Kubernetes Local

#### Cách 1: Sử dụng Docker Desktop (Khuyến nghị)

1. **Cài đặt Docker Desktop**
   - Tải từ: https://www.docker.com/products/docker-desktop/
   - Cài đặt và khởi động Docker Desktop

2. **Bật Kubernetes trong Docker Desktop**
   - Mở Docker Desktop
   - Vào Settings (⚙️)
   - Chọn "Kubernetes" trong menu bên trái
   - Tick vào "Enable Kubernetes"
   - Chọn "Apply & Restart"
   - Chờ Docker Desktop khởi động lại

3. **Kiểm tra cài đặt**
   ```powershell
   # Chạy script kiểm tra
   .\setup-k8s-local.ps1
   
   # Hoặc kiểm tra thủ công
   kubectl cluster-info
   kubectl get nodes
   ```

#### Cách 2: Sử dụng Minikube

1. **Cài đặt Minikube**
   ```powershell
   # Sử dụng Chocolatey
   choco install minikube
   
   # Hoặc tải trực tiếp
   # https://minikube.sigs.k8s.io/docs/start/
   ```

2. **Khởi động Minikube**
   ```powershell
   minikube start
   minikube status
   ```

3. **Cấu hình kubectl**
   ```powershell
   kubectl config use-context minikube
   ```

### Bước 2: Cài đặt kubectl

#### Windows
```powershell
# Sử dụng Chocolatey
choco install kubernetes-cli

# Hoặc tải trực tiếp
# https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/
```

#### Kiểm tra cài đặt
```powershell
kubectl version --client
```

### Bước 3: Triển khai ứng dụng

#### Sử dụng script tự động
```powershell
# Thiết lập cluster
.\setup-k8s-local.ps1

# Triển khai ứng dụng
.\deploy-k8s-local.ps1
```

#### Triển khai thủ công
```powershell
# 1. Build Docker image
docker build -t url-shortener:latest .

# 2. Load image vào cluster (nếu sử dụng minikube)
minikube image load url-shortener:latest

# 3. Triển khai
kubectl apply -f k8s/
```

### Bước 4: Kiểm tra triển khai

```powershell
# Xem pods
kubectl get pods -n url-shortener

# Xem services
kubectl get services -n url-shortener

# Xem logs
kubectl logs -f deployment/url-shortener -n url-shortener

# Port forward để truy cập
kubectl port-forward -n url-shortener service/url-shortener-service 8080:80
```

## Troubleshooting

### Lỗi thường gặp

#### 1. "kubectl: command not found"
```powershell
# Cài đặt kubectl
choco install kubernetes-cli
```

#### 2. "The connection to the server was refused"
```powershell
# Kiểm tra Docker Desktop
docker ps

# Bật Kubernetes trong Docker Desktop
# Settings > Kubernetes > Enable Kubernetes
```

#### 3. "no context is set"
```powershell
# Kiểm tra contexts
kubectl config get-contexts

# Chọn context
kubectl config use-context docker-desktop
# hoặc
kubectl config use-context minikube
```

#### 4. "ImagePullBackOff" hoặc "ErrImagePull"
```powershell
# Build image local
docker build -t url-shortener:latest .

# Load vào minikube (nếu sử dụng minikube)
minikube image load url-shortener:latest

# Hoặc sử dụng Docker Desktop (không cần load)
```

#### 5. Pod không start
```powershell
# Xem chi tiết pod
kubectl describe pod <pod-name> -n url-shortener

# Xem logs
kubectl logs <pod-name> -n url-shortener

# Kiểm tra events
kubectl get events -n url-shortener --sort-by='.lastTimestamp'
```

### Kiểm tra tài nguyên

```powershell
# Kiểm tra nodes
kubectl get nodes

# Kiểm tra tài nguyên
kubectl top nodes
kubectl top pods -n url-shortener

# Kiểm tra storage
kubectl get pv,pvc -n url-shortener
```

## Cấu hình nâng cao

### Ingress Controller

Nếu muốn sử dụng Ingress:

```powershell
# Cài đặt NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

# Kiểm tra
kubectl get pods -n ingress-nginx
```

### Monitoring

```powershell
# Cài đặt metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Kiểm tra
kubectl top nodes
```

## Tài liệu tham khảo

- [Docker Desktop Kubernetes](https://docs.docker.com/desktop/kubernetes/)
- [Minikube](https://minikube.sigs.k8s.io/docs/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

## Lưu ý

- Docker Desktop yêu cầu ít nhất 4GB RAM
- Minikube yêu cầu ít nhất 2GB RAM
- Đảm bảo có đủ dung lượng ổ cứng (ít nhất 10GB)
- Trên Windows, có thể cần bật Hyper-V hoặc VirtualBox
