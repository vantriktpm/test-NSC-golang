# Kafka-based Inventory Management System

## Kiến trúc tổng quan

```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │     (Nginx)     │
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
                    │   Kafka Cluster │
                    │   (3 Brokers)   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │Inventory│          │Inventory│          │Inventory│
   │Consumer │          │Consumer │          │Consumer │
   │Group 1  │          │Group 2  │          │Group 3  │
   └────┬────┘          └────┬────┘          └────┬────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL   │
                    │   (Inventory   │
                    │    Database)   │
                    └────────────────┘
```

## Kafka Topics Design

### 1. Inventory Events Topic
```yaml
Topic: inventory-events
Partitions: 12 (4 partitions per product category)
Replication Factor: 3
Retention: 7 days
```

**Message Schema:**
```json
{
  "eventId": "uuid",
  "eventType": "INVENTORY_CHECK|INVENTORY_RESERVE|INVENTORY_CONFIRM|INVENTORY_RELEASE",
  "productId": "string",
  "userId": "string",
  "quantity": "number",
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "metadata": {
    "userAgent": "string",
    "ipAddress": "string",
    "sessionId": "string"
  }
}
```

### 2. Inventory State Topic
```yaml
Topic: inventory-state
Partitions: 12
Replication Factor: 3
Retention: 30 days
Compaction: true
```

**Message Schema:**
```json
{
  "productId": "string",
  "availableStock": "number",
  "reservedStock": "number",
  "totalStock": "number",
  "lastUpdated": "ISO8601",
  "version": "number"
}
```

## Luồng xử lý mới

### 1. User Request Flow
```
User Request → App Server → Kafka Producer → inventory-events topic
     ↓
Response (Async) ← WebSocket ← Kafka Consumer ← inventory-state topic
```

### 2. Inventory Processing Flow
```
Kafka Event → Inventory Consumer → Process Logic → Update Database
     ↓
Publish State Update → inventory-state topic → Notify All Clients
```

## Các Kafka Consumer Groups

### 1. Inventory Processor Group
- **Function**: Xử lý inventory events và cập nhật database
- **Instances**: 3 consumers (1 per server)
- **Processing**: Sequential processing per productId để đảm bảo consistency

### 2. State Publisher Group  
- **Function**: Publish inventory state changes
- **Instances**: 3 consumers
- **Processing**: Real-time state broadcasting

### 3. Analytics Group
- **Function**: Collect analytics data
- **Instances**: 2 consumers
- **Processing**: Async analytics processing

## Database Schema Updates

```sql
-- Products table
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    total_stock INTEGER NOT NULL DEFAULT 0,
    available_stock INTEGER NOT NULL DEFAULT 0,
    reserved_stock INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Inventory events table (for audit trail)
CREATE TABLE inventory_events (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    user_id UUID,
    quantity INTEGER NOT NULL,
    correlation_id UUID,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- User reservations table
CREATE TABLE user_reservations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    correlation_id UUID NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_products_available_stock ON products(available_stock);
CREATE INDEX idx_inventory_events_product_id ON inventory_events(product_id);
CREATE INDEX idx_user_reservations_user_product ON user_reservations(user_id, product_id);
CREATE INDEX idx_user_reservations_expires_at ON user_reservations(expires_at);
```

## Performance Benefits

### Trước khi tối ưu (Redis + Database):
- **Throughput**: ~1000 requests/second
- **Latency**: 50-200ms per request
- **Database Load**: High (mỗi request = 2-3 queries)
- **Scalability**: Limited by database connections

### Sau khi tối ưu (Kafka-based):
- **Throughput**: ~100,000+ requests/second
- **Latency**: 5-20ms per request (async response)
- **Database Load**: Low (batch processing)
- **Scalability**: Horizontal scaling với Kafka partitions

## Implementation Strategy

### Phase 1: Core Kafka Infrastructure
1. Setup Kafka cluster (3 brokers)
2. Create topics với proper partitioning
3. Implement basic producers và consumers

### Phase 2: Inventory Service
1. Implement inventory event processing
2. Database integration
3. State management

### Phase 3: API Integration
1. Update existing APIs để sử dụng Kafka
2. WebSocket integration cho real-time updates
3. Error handling và retry mechanisms

### Phase 4: Monitoring & Optimization
1. Kafka metrics monitoring
2. Performance tuning
3. Load testing với hàng triệu requests
